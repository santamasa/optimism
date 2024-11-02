package proofs

import (
	"bytes"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"testing"
	"time"

	actionsHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/actions/proofs/helpers"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	dtest "github.com/ethereum-optimism/optimism/op-node/rollup/derive/test"
	"github.com/stretchr/testify/require"
)

func Test_ProgramAction_BigChannel(gt *testing.T) {

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

		parentHash := env.Sequencer.RollupCfg.Genesis.L2.Hash
		parentNumber := big.NewInt(0)
		for uint64(hugeChannelOut.RLPLength()) < env.Sd.ChainSpec.MaxRLPBytesPerChannel(uint64(time.Now().Unix())) {
			block := dtest.HighlyCompressible2BlockWithChainIdAndTime(rng, 1000, env.Sequencer.RollupCfg.L2ChainID, time.Time{})
			bHeader := block.Header()
			bHeader.Number = new(big.Int).Add(parentNumber, big.NewInt(1))
			bHeader.ParentHash = parentHash
			block = block.WithSeal(bHeader)
			parentNumber = bHeader.Number
			parentHash = bHeader.Root
			t.Log(block.Number())
			_, err := hugeChannelOut.AddBlock(env.Sequencer.RollupCfg, block)
			if err != nil {
				t.Fatal(err)
			}
		}
		hugeChannelOut.Close()

		t.Log(hugeChannelOut.RLPLength(), hugeChannelOut.ReadyBytes())

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
			for {
				// Collect the output frame
				data := new(bytes.Buffer)
				data.WriteByte(derive.DerivationVersion0)
				_, err := hugeChannelOut.OutputFrame(data, 100_000)
				if err == io.EOF {
					break
				} else if err != nil {
					t.Fatal(err)
				}
				env.Batcher.ActL2BatchSubmitRaw(t, data.Bytes())
				includeBatchTx()
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
	}
}
