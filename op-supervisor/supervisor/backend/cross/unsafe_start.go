package cross

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-supervisor/supervisor/types"
)

type UnsafeStartDeps interface {
	Check(chain types.ChainID, blockNum uint64, logIdx uint32, logHash common.Hash) (includedIn types.BlockSeal, err error)

	IsCrossUnsafe(chainID types.ChainID, block eth.BlockID) error
}

// CrossUnsafeHazards checks if the given messages all exist and pass invariants.
// It returns a hazard-set: if any intra-block messaging happened,
// these hazard blocks have to be verified.
func CrossUnsafeHazards(d UnsafeStartDeps, chainID types.ChainID,
	candidate types.BlockSeal, execMsgs []*types.ExecutingMessage) (hazards map[types.ChainIndex]types.BlockSeal, err error) {

	hazards = make(map[types.ChainIndex]types.BlockSeal)

	// Warning for future: If we have sub-second distinct blocks (different block number),
	// we need to increase precision on the above timestamp invariant.
	// Otherwise a local block can depend on a future local block of the same chain,
	// simply by pulling in a block of another chain,
	// which then depends on a block of the original chain,
	// all with the same timestamp, without message cycles.

	// check all executing messages
	for _, msg := range execMsgs {
		execChainID := types.ChainIDFromUInt64(uint64(msg.Chain)) // TODO(#11105): translate chain index to chain ID
		if msg.Timestamp < candidate.Timestamp {
			// If timestamp is older: invariant ensures non-cyclic ordering relative to other messages.
			// Check that the block that they are included in is cross-safe already.
			includedIn, err := d.Check(execChainID, msg.BlockNum, msg.LogIdx, msg.Hash)
			if err != nil {
				// TODO
			}
			if err := d.IsCrossUnsafe(execChainID, includedIn.ID()); err != nil {

			}
		} else if msg.Timestamp == candidate.Timestamp {
			// If timestamp is equal: we have to inspect ordering of individual
			// log events to ensure non-cyclic cross-chain message ordering.
			// And since we may have back-and-forth messaging, we cannot wait till the initiating side is cross-unsafe.
			// Thus check that it was included in a local-unsafe block,
			// and then proceed with transitive block checks,
			// to ensure the local block we depend on is becoming cross-unsafe also.
			includedIn, err := d.Check(execChainID, msg.BlockNum, msg.LogIdx, msg.Hash)
			if err != nil {
				// TODO
			}

			// TODO
			if existing, ok := hazards[msg.Chain]; ok {
				if existing != includedIn {
					return nil, fmt.Errorf("found dependency on %s (chain %d), but already depend on %s", includedIn, execChainID, chainID)
				}
			} else {
				// Mark it as hazard block
				hazards[msg.Chain] = includedIn
			}
		} else {
			// Timestamp invariant is broken: executing message tries to execute future block.
			// The predeploy inbox contract should not have allowed this executing message through.
			return nil, fmt.Errorf("executing message %s in %s breaks timestamp invariant", msg, candidate)
		}
	}
	return hazards, nil
}
