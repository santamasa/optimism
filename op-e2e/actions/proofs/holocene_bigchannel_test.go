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
			block := dtest.HighlyCompressible2BlockWithChainIdAndTime(rng, 1000, env.Sequencer.RollupCfg.L2ChainID, time.Now())
			bHeader := block.Header()
			bHeader.Number = new(big.Int).Add(parentNumber, big.NewInt(1))
			bHeader.ParentHash = parentHash
			block = block.WithSeal(bHeader)
			parentNumber = bHeader.Number
			parentHash = bHeader.Root
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

		for {
			// Collect the output frames, submit and include them.
			data := new(bytes.Buffer)
			data.WriteByte(derive.DerivationVersion0)
			_, err := hugeChannelOut.OutputFrame(data, 100_000)
			if err == io.EOF {
				env.Batcher.ActL2BatchSubmitRaw(t, data.Bytes())
				includeBatchTx()
				break
			} else if err != nil {
				t.Fatal(err)
			}
			env.Batcher.ActL2BatchSubmitRaw(t, data.Bytes())
			includeBatchTx()
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
