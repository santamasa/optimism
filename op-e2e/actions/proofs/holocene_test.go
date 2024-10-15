package proofs

import (
	"fmt"
	"testing"

	actionsHelpers "github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/actions/proofs/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils"
	"github.com/ethereum-optimism/optimism/op-program/client/claim"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

type ordering []uint

// badOrderings is a list of orderings for
// a three frame channel which are all
// valid pre-Holocene but are invalid
// post-Holocene.
// The correct ordering is {0,1,2}
var badOrderings = []ordering{
	{0, 2, 1},
	{2, 1, 0},
	{0, 1, 0, 2},
}

func Test_ProgramAction_HoloceneFrameRules(gt *testing.T) {
	matrix := helpers.NewMatrix[ordering]()
	defer matrix.Run(gt)

	for _, ordering := range badOrderings {
		matrix.AddTestCase(
			fmt.Sprintf("HonestClaim-%v", ordering),
			ordering,
			helpers.NewForkMatrix(helpers.Granite, helpers.LatestFork),
			runHoloceneFrameTest,
			helpers.ExpectNoError(),
		)
		matrix.AddTestCase(
			fmt.Sprintf("JunkClaim-%v", ordering),
			ordering,
			helpers.NewForkMatrix(helpers.Granite, helpers.LatestFork),
			runHoloceneFrameTest,
			helpers.ExpectError(claim.ErrClaimNotValid),
			helpers.WithL2Claim(common.HexToHash("0xdeadbeef")),
		)
	}
}

func runHoloceneFrameTest(gt *testing.T, testCfg *helpers.TestCfg[ordering]) {
	t := actionsHelpers.NewDefaultTesting(gt)
	tp := helpers.NewTestParams(func(tp *e2eutils.TestParams) {
		// Set the channel timeout to 10 blocks, 12x lower than the sequencing window.
		tp.ChannelTimeout = 10
	})
	env := helpers.NewL2FaultProofEnv(t, testCfg, tp, helpers.NewBatcherCfg())

	includeBatchTx := func() {
		// Include the last transaction submitted by the batcher.
		env.Miner.ActL1StartBlock(12)(t)
		env.Miner.ActL1IncludeTxByHash(env.Batcher.LastSubmitted.Hash())(t)
		env.Miner.ActL1EndBlock(t)

		// Finalize the block with the first channel frame on L1.
		env.Miner.ActL1SafeNext(t)
		env.Miner.ActL1FinalizeNext(t)

		// Instruct the sequencer to derive the L2 chain from the data on L1 that the batcher just posted.
		env.Sequencer.ActL1HeadSignal(t)
		env.Sequencer.ActL2PipelineFull(t)
	}

	const NumL2Blocks = 15
	// Build NumL2Blocks empty blocks on L2
	for i := 0; i < NumL2Blocks; i++ {
		env.Sequencer.ActL2StartBlock(t)
		env.Sequencer.ActL2EndBlock(t)
	}

	// Buffer the first third of L2 blocks in the batcher, and submit it.
	for i := 0; i < NumL2Blocks/3; i++ {
		env.Batcher.ActL2BatchBuffer(t)
	}
	orderedFrames := [][]byte{env.Batcher.ReadNextOutputFrame(t)}

	// Buffer the second third of L2 blocks in the batcher.
	for i := 0; i < NumL2Blocks/2; i++ {
		env.Batcher.ActL2BatchBuffer(t)
	}
	orderedFrames = append(orderedFrames, env.Batcher.ReadNextOutputFrame(t))

	// Buffer the final third of L2 blocks in the batcher.
	for i := 0; i < NumL2Blocks/2; i++ {
		env.Batcher.ActL2BatchBuffer(t)
	}
	env.Batcher.ActL2ChannelClose(t)
	orderedFrames = append(orderedFrames, env.Batcher.ReadNextOutputFrame(t))

	// Submit frames out of order
	for _, j := range testCfg.Custom {
		env.Batcher.ActL2BatchSubmitRaw(t, orderedFrames[j])
		includeBatchTx()
	}

	l2SafeHead := env.Engine.L2Chain().CurrentSafeBlock()

	if testCfg.Hardfork.Precedence < helpers.Holocene.Precedence {
		// The safe head should have still advanced, since Holocene rules are not activated yet
		// and the entire channel was submitted
		require.Equal(t, uint64(NumL2Blocks), l2SafeHead.Number.Uint64())
	} else {
		// The safe head should not have advanced, since the Holocene rules were
		// violated (no contiguous and complete run of frames from the channel)
		t.Log("Holocene derivation rules not yet implemented")
		// require.Equal(t, uint64(0), l2SafeHead.Number.Uint64()) // TODO activate this line
	}

	// Run the FPP on L2 block # NumL2Blocks.
	env.RunFaultProofProgram(t, NumL2Blocks, testCfg.CheckResult, testCfg.InputParams...)
}
