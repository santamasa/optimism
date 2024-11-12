package engine

import (
	"time"

	"github.com/ethereum-optimism/optimism/op-service/eth"
)

type PayloadSuccessEvent struct {
	// if payload should be promoted to (local) safe (must also be pending safe, see DerivedFrom)
	Concluding bool
	// payload is promoted to pending-safe if non-zero
	DerivedFrom  eth.L1BlockRef
	BuildStarted time.Time

	Envelope *eth.ExecutionPayloadEnvelope
	Ref      eth.L2BlockRef
}

func (ev PayloadSuccessEvent) String() string {
	return "payload-success"
}

func (eq *EngDeriver) onPayloadSuccess(ev PayloadSuccessEvent) {
	eq.emitter.Emit(PromoteUnsafeEvent{Ref: ev.Ref})

	// If derived from L1, then it can be considered (pending) safe
	if ev.DerivedFrom != (eth.L1BlockRef{}) {
		eq.emitter.Emit(PromotePendingSafeEvent{
			Ref:         ev.Ref,
			Concluding:  ev.Concluding,
			DerivedFrom: ev.DerivedFrom,
		})
	}

	elapsed := time.Since(ev.BuildStarted)
	payload := ev.Envelope.ExecutionPayload
	eq.log.Info("Inserted block", "hash", payload.BlockHash, "number", uint64(payload.BlockNumber),
		"state_root", payload.StateRoot, "timestamp", uint64(payload.Timestamp), "parent", payload.ParentHash,
		"prev_randao", payload.PrevRandao, "fee_recipient", payload.FeeRecipient,
		"txs", len(payload.Transactions), "concluding", ev.Concluding, "derived_from", ev.DerivedFrom,
		"elapsed", elapsed, "mgas", float64(payload.GasUsed)/1000000,
		"mgasps", float64(payload.GasUsed)*1000/float64(elapsed))

	eq.emitter.Emit(TryUpdateEngineEvent{})
}
