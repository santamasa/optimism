package proofs

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	actionsHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/actions/proofs/helpers"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-program/client/claim"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func Test_ProgramAction_BigChannel(gt *testing.T) {

	type testCase struct {
		name               string
		disableCompression bool
	}

	var testCases = []testCase{
		{name: "case-0"},
		{name: "case-1", disableCompression: true}}

	runHoloceneDerivationTest := func(gt *testing.T, testCfg *helpers.TestCfg[testCase]) {
		t := actionsHelpers.NewDefaultTesting(gt)
		batcherConfig := helpers.NewBatcherCfg()
		batcherConfig.GarbageCfg = &actionsHelpers.GarbageChannelCfg{
			IgnoreMaxRLPBytesPerChannel: true,
			DisableCompression:          testCfg.Custom.disableCompression,
		}
		env := helpers.NewL2FaultProofEnv(t, testCfg, helpers.NewTestParams(), batcherConfig)

		// build some l1 blocks so that we don't hit sequencer drift problems
		for i := 0; i < 200; i++ {
			env.Miner.ActEmptyBlock(t)
		}

		hugeChannelOut, _ := actionsHelpers.NewGarbageChannelOut(batcherConfig.GarbageCfg)

		parentTime := env.Sequencer.RollupCfg.Genesis.L2Time
		blockTime := env.Sequencer.RollupCfg.BlockTime
		for uint64(hugeChannelOut.RLPLength()) < env.Sd.ChainSpec.MaxRLPBytesPerChannel(parentTime+blockTime) {
			env.Sequencer.ActL2StartBlock(t)
			for i := 0; i < 2; i++ {
				env.Alice.L2.ActResetTxOpts(t)
				env.Alice.L2.ActSetTxToAddr(&env.Dp.Addresses.Bob)(t)
				env.Alice.L2.ActSetTxCalldata(bytes.Repeat([]byte{1}, 130_000))(t)
				env.Alice.L2.ActMakeTx(t)
				env.Engine.ActL2IncludeTx(env.Alice.Address())(t)
			}

			env.Sequencer.ActL2EndBlock(t)

			unsafeHeadNumber := env.Sequencer.L2Unsafe().Number
			t.Log("unsafe l2 head number", unsafeHeadNumber)

			block := env.Engine.Eth.BlockChain().GetBlockByNumber(unsafeHeadNumber)
			_, err := hugeChannelOut.AddBlock(env.Sequencer.RollupCfg, block)
			if err != nil {
				t.Fatal(err)
			}
		}
		err := hugeChannelOut.Close()
		require.NoError(t, err)

		t.Log("closed channel", "rlp_length", hugeChannelOut.RLPLength(), "ready_bytes", hugeChannelOut.ReadyBytes())

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

		holoceneExpectations := holoceneExpectations{}
		if testCfg.Custom.disableCompression {
			holoceneExpectations.safeHeadHolocene = 0 // entire channel dropped because the compressed
			holoceneExpectations.safeHeadPreHolocene = env.Sequencer.L2Unsafe().Number - 1
		} else {
			// Because the channel will be _clipped_ to max_rlp_bytes_per_channel, the safe
			// head is expected to move up to but not including the last block in the channel.
			holoceneExpectations.safeHeadHolocene = env.Sequencer.L2Unsafe().Number - 1
			holoceneExpectations.safeHeadPreHolocene = env.Sequencer.L2Unsafe().Number - 1
		}

		holoceneExpectations.RequireExpectedProgress(t, l2SafeHead, testCfg.Hardfork.Precedence < helpers.Holocene.Precedence, env.Engine)

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
