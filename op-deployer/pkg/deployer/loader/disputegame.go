package loader

import (
	"context"
	"fmt"

	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/opcm"
	"github.com/ethereum-optimism/optimism/packages/contracts-bedrock/snapshots"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

func LoadDisputeGameInputs(ctx context.Context, l1Client ContractCaller, source common.Address, permissioned bool) (opcm.DeployDisputeGameInput, error) {
	var gameABI *abi.ABI
	if permissioned {
		gameABI = snapshots.LoadPermissionedDisputeGameABI()
	} else {
		gameABI = snapshots.LoadFaultDisputeGameABI()
	}
	caller := NewContractAPI(l1Client, gameABI, source)
	vm, err := caller.GetAddress(ctx, "vm")
	if err != nil {
		return opcm.DeployDisputeGameInput{}, fmt.Errorf("failed to load vm from %v: %w", source, err)
	}
	delayedWETHProxy, err := caller.GetAddress(ctx, "weth")
	if err != nil {
		return opcm.DeployDisputeGameInput{}, fmt.Errorf("failed to load DelayedWETH proxy from %v: %w", source, err)
	}
	anchorStateRegistryProxy, err := caller.GetAddress(ctx, "anchorStateRegistry")
	if err != nil {
		return opcm.DeployDisputeGameInput{}, fmt.Errorf("failed to load AnchorStateRegistry proxy from %v: %w", source, err)
	}
	l2ChainID, err := caller.GetBigInt(ctx, "l2ChainId")
	if err != nil {
		return opcm.DeployDisputeGameInput{}, fmt.Errorf("failed to load L2 chain ID from %v: %w", source, err)
	}
	absolutePrestate, err := caller.GetHash(ctx, "absolutePrestate")
	if err != nil {
		return opcm.DeployDisputeGameInput{}, fmt.Errorf("failed to load absolute prestate from %v: %w", source, err)
	}
	gameType, err := caller.GetUint32(ctx, "gameType")
	if err != nil {
		return opcm.DeployDisputeGameInput{}, fmt.Errorf("failed to load game type from %v: %w", source, err)
	}
	maxGameDepth, err := caller.GetBigInt(ctx, "maxGameDepth")
	if err != nil {
		return opcm.DeployDisputeGameInput{}, fmt.Errorf("failed to load max game depth from %v: %w", source, err)
	}
	splitDepth, err := caller.GetBigInt(ctx, "splitDepth")
	if err != nil {
		return opcm.DeployDisputeGameInput{}, fmt.Errorf("failed to load split depth from %v: %w", source, err)
	}
	clockExtension, err := caller.GetUint64(ctx, "clockExtension")
	if err != nil {
		return opcm.DeployDisputeGameInput{}, fmt.Errorf("failed to load clock extension from %v: %w", source, err)
	}
	maxClockDuration, err := caller.GetUint64(ctx, "maxClockDuration")
	if err != nil {
		return opcm.DeployDisputeGameInput{}, fmt.Errorf("failed to load max clock duration from %v: %w", source, err)
	}
	var proposer common.Address
	var challenger common.Address
	gameKind := "FaultDisputeGame"
	if permissioned {
		gameKind = "PermissionedDisputeGame"
		proposer, err = caller.GetAddress(ctx, "proposer")
		if err != nil {
			return opcm.DeployDisputeGameInput{}, fmt.Errorf("failed to load proposer from %v: %w", source, err)
		}
		challenger, err = caller.GetAddress(ctx, "challenger")
		if err != nil {
			return opcm.DeployDisputeGameInput{}, fmt.Errorf("failed to load challenger from %v: %w", source, err)
		}
	}
	return opcm.DeployDisputeGameInput{
		FpVm:                     vm,
		GameKind:                 gameKind,
		GameType:                 gameType,
		AbsolutePrestate:         absolutePrestate,
		MaxGameDepth:             maxGameDepth.Uint64(),
		SplitDepth:               splitDepth.Uint64(),
		ClockExtension:           clockExtension,
		MaxClockDuration:         maxClockDuration,
		DelayedWethProxy:         delayedWETHProxy,
		AnchorStateRegistryProxy: anchorStateRegistryProxy,
		L2ChainId:                l2ChainID.Uint64(),
		Proposer:                 proposer,
		Challenger:               challenger,
	}, nil
}
