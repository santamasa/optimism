package proofs

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"

	actionsHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/actions/proofs/helpers"
	dtest "github.com/ethereum-optimism/optimism/op-node/rollup/derive/test"
	"github.com/ethereum-optimism/optimism/op-program/client/claim"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

type holoceneExpectations struct {
	safeHeadPreHolocene uint64
	safeHeadHolocene    uint64
}

func (h holoceneExpectations) RequireExpectedProgress(t actionsHelpers.StatefulTesting, actualSafeHead eth.L2BlockRef, isHolocene bool, engine *actionsHelpers.L2Engine) {
	if isHolocene {
		require.Equal(t, h.safeHeadPreHolocene, actualSafeHead.Number)
		expectedHash := engine.L2Chain().GetBlockByNumber(h.safeHeadPreHolocene).Hash()
		require.Equal(t, expectedHash, actualSafeHead.Hash)
	} else {
		require.Equal(t, h.safeHeadHolocene, actualSafeHead.Number)
		expectedHash := engine.L2Chain().GetBlockByNumber(h.safeHeadHolocene).Hash()
		require.Equal(t, expectedHash, actualSafeHead.Hash)
	}
}
func Test_ProgramAction_HoloceneFrames(gt *testing.T) {

	type testCase struct {
		name   string
		frames []uint
		holoceneExpectations
	}

	// An ordered list of frames to read from the channel and submit
	// on L1. We expect a different progression of the safe head under Holocene
	// derivation rules, compared with pre Holocene.
	var testCases = []testCase{
		// Standard frame submission,
		{name: "case-0", frames: []uint{0, 1, 2},
			holoceneExpectations: holoceneExpectations{
				safeHeadPreHolocene: 3,
				safeHeadHolocene:    3},
		},

		// Non-standard frame submission
		{name: "case-1a", frames: []uint{2, 1, 0},
			holoceneExpectations: holoceneExpectations{
				safeHeadPreHolocene: 3, // frames are buffered, so ordering does not matter
				safeHeadHolocene:    0, // non-first frames will be dropped b/c it is the first seen with that channel Id. The safe head won't move until the channel is closed/completed.
			},
		},
		{name: "case-1b", frames: []uint{0, 1, 0, 2},
			holoceneExpectations: holoceneExpectations{
				safeHeadPreHolocene: 3, // frames are buffered, so ordering does not matter
				safeHeadHolocene:    0, // non-first frames will be dropped b/c it is the first seen with that channel Id. The safe head won't move until the channel is closed/completed.
			},
		},
		{name: "case-1c", frames: []uint{0, 1, 1, 2},
			holoceneExpectations: holoceneExpectations{
				safeHeadPreHolocene: 3, // frames are buffered, so ordering does not matter
				safeHeadHolocene:    3, // non-contiguous frames are dropped. So this reduces to case-0.
			},
		},
	}

	veryCompressibleCalldata := make([]byte, 49_000)
	for i := 0; i < len(veryCompressibleCalldata); i++ {
		veryCompressibleCalldata[i] = 1
	}
	runHoloceneDerivationTest := func(gt *testing.T, testCfg *helpers.TestCfg[testCase]) {
		t := actionsHelpers.NewDefaultTesting(gt)
		batcherConfig := helpers.NewBatcherCfg()
		batcherConfig.GarbageCfg = &actionsHelpers.GarbageChannelCfg{IgnoreMaxRLPBytesPerChannel: true}
		env := helpers.NewL2FaultProofEnv(t, testCfg, helpers.NewTestParams(), batcherConfig)

		k := 1000
		blocks := make([]uint, k)
		for i := 0; i < k; i++ {
			blocks[i] = uint(i) + 1
		}

		hugeChannelOut, _ := actionsHelpers.NewGarbageChannelOut(&actionsHelpers.GarbageChannelCfg{IgnoreMaxRLPBytesPerChannel: true})

		rng := rand.New(rand.NewSource(1234))

		for uint64(hugeChannelOut.ReadyBytes()) < env.Sd.ChainSpec.MaxRLPBytesPerChannel(uint64(time.Now().Unix())) {
			block := dtest.RandomL2BlockWithChainId(rng, 1000, env.Sequencer.RollupCfg.L2ChainID)
			_, err := hugeChannelOut.AddBlock(env.Sequencer.RollupCfg, block)
			if err != nil {
				t.Fatal(err)
			}
		}

		t.Log(hugeChannelOut.ReadyBytes())
		includeBatchTx := func() {
			// Include the last transaction submitted by the batcher.
			env.Miner.ActL1StartBlock(12)(t)
			env.Miner.ActL1IncludeTxByHash(env.Batcher.LastSubmitted.Hash())(t)
			env.Miner.ActL1EndBlock(t)

			// Finalize the block with the first channel frame on L1.
			env.Miner.ActL1SafeNext(t)
			env.Miner.ActL1FinalizeNext(t)
		}
		if false { // TODO replace with a switch on the test case

			aliceAddress := env.Alice.Address()
			targetHeadNumber := k
			for env.Engine.L2Chain().CurrentBlock().Number.Uint64() < uint64(targetHeadNumber) {
				env.Sequencer.ActL2StartBlock(t)

				// alice makes several L2 txs, sequencer includes them
				for i := 0; i < 100; i++ {
					env.Alice.L2.ActResetTxOpts(t)
					env.Alice.L2.ActSetTxCalldata(veryCompressibleCalldata)(t)
					env.Alice.L2.ActMakeTx(t)
					env.Engine.ActL2IncludeTx(aliceAddress)(t)
				}
				env.Alice.L2.ActResetTxOpts(t)
				env.Alice.L2.ActSetTxToAddr(&env.Dp.Addresses.Bob)
				env.Alice.L2.ActMakeTx(t)
				env.Engine.ActL2IncludeTx(env.Alice.Address())(t)
				env.Sequencer.ActL2EndBlock(t)
			}

			// Build up a local list of frames
			orderedFrames := make([][]byte, 0, len(testCfg.Custom.frames))
			// Buffer the blocks in the batcher and populate orderedFrames list
			env.Batcher.ActCreateChannel(t, false)
			for i, blockNum := range blocks {
				env.Batcher.ActAddBlockByNumber(t, int64(blockNum), actionsHelpers.BlockLogger(t))
				if i == len(blocks)-1 {
					env.Batcher.ActL2ChannelClose(t)
				}
				frame := env.Batcher.ReadNextOutputFrame(t)
				require.NotEmpty(t, frame, "frame %d", i)
				orderedFrames = append(orderedFrames, frame)
			}

			// Submit frames in specified order order
			for _, j := range testCfg.Custom.frames {
				env.Batcher.ActL2BatchSubmitRaw(t, orderedFrames[j])
				includeBatchTx()
			}
		} else {
			frames := make([][]byte, 0, 1000)
			for {
				frame := new(bytes.Buffer)
				_, err := hugeChannelOut.OutputFrame(frame, 100_000)
				if err == io.EOF {
					break
				}
				frames = append(frames, frame.Bytes())
			}

			for _, frame := range frames {
				env.Miner.ActL1StartBlock(12)(t)
				env.Batcher.ActL2BatchSubmitRaw(t, frame)
				env.Miner.ActL1IncludeTxByHash(env.Batcher.LastSubmitted.Hash())(t)
				env.Miner.ActL1EndBlock(t)
			}

		}

		// Instruct the sequencer to derive the L2 chain from the data on L1 that the batcher just posted.
		env.Sequencer.ActL1HeadSignal(t)
		env.Sequencer.ActL2PipelineFull(t)

		l2SafeHead := env.Sequencer.L2Safe()

		testCfg.Custom.RequireExpectedProgress(t, l2SafeHead, testCfg.Hardfork.Precedence < helpers.Holocene.Precedence, env.Engine)

		t.Log("Safe head progressed as expected", "l2SafeHeadNumber", l2SafeHead.Number)

		if safeHeadNumber := l2SafeHead.Number; safeHeadNumber > 0 {
			env.RunFaultProofProgram(t, safeHeadNumber, testCfg.CheckResult, testCfg.InputParams...)
		}
	}

	matrix := helpers.NewMatrix[testCase]()
	defer matrix.Run(gt)

	for _, ordering := range testCases {
		matrix.AddTestCase(
			fmt.Sprintf("HonestClaim-%s", ordering.name),
			ordering,
			helpers.NewForkMatrix(helpers.Granite, helpers.LatestFork),
			runHoloceneDerivationTest,
			helpers.ExpectNoError(),
		)
		matrix.AddTestCase(
			fmt.Sprintf("JunkClaim-%s", ordering.name),
			ordering,
			helpers.NewForkMatrix(helpers.Granite, helpers.LatestFork),
			runHoloceneDerivationTest,
			helpers.ExpectError(claim.ErrClaimNotValid),
			helpers.WithL2Claim(common.HexToHash("0xdeadbeef")),
		)
	}
}