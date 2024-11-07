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

		hugeChannelOut, err := actionsHelpers.NewGarbageChannelOut(batcherConfig.GarbageCfg)
		require.NoError(t, err)

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
		err = hugeChannelOut.Close()
		require.NoError(t, err)

		t.Log("closed channel", "rlp_length", hugeChannelOut.RLPLength(), "ready_bytes", hugeChannelOut.ReadyBytes())

		frames := make([][]byte, 0)
		for {
			// Collect the output frames and submit them.
			data := new(bytes.Buffer)
			data.WriteByte(derive.DerivationVersion0)
			_, err := hugeChannelOut.OutputFrame(data, 130_000) // close to max blob size
			// The channel must be > 100MB compressed to be impossible to get on chain

			if err == io.EOF {
				frames = append(frames, data.Bytes())
				break
			} else if err != nil {
				t.Fatal(err)
			}
			frames = append(frames, data.Bytes())
		}

		// To avoid the channel timing out, we need to get it on chain within
		// CHANNEL_TIMEOUT which is 50 L1 blocks when Granite is activated.
		// 100MB / 50 blocks = 2MB per block
		// This exceeds the capacity of L1.
		// We can use 6 blobs per block at 130KB per blob, which is 780KB per block.
		// Only with the longer term limit of 16 blobs per block could we get up to 2MB per block.
		// Or we can use calldata, which could mean up to 7.5MB per block if the data
		// is all zeros. It may also be easier to modify the limits for calldata in the test environment.
		for _, frame := range frames {
			env.Miner.ActL1StartBlock(12)(t)
			for i := 0; i < 16; i++ {
				env.Batcher.ActL2BatchSubmitRaw(t, frame)
				env.Miner.ActL1IncludeTxByHash(env.Batcher.LastSubmitted.Hash())(t)
			}
			env.Miner.ActL1EndBlock(t)
		}

		// Instruct the sequencer to derive the L2 chain from the data on L1 that the batcher just posted.
		env.Sequencer.ActL1HeadSignal(t)
		env.Sequencer.ActL2PipelineFull(t)

		l2SafeHead := env.Sequencer.L2Safe()

		holoceneExpectations := holoceneExpectations{}
		if testCfg.Custom.disableCompression {
			holoceneExpectations.safeHeadHolocene = 0 // entire channel dropped because the compressed size is too big
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
