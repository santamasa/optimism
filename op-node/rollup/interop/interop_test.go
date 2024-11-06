package interop

import (
	"context"
	"math/big"
	"math/rand" // nosemgrep
	"testing"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/engine"
	"github.com/ethereum-optimism/optimism/op-node/rollup/finality"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/testutils"
	supervisortypes "github.com/ethereum-optimism/optimism/op-supervisor/supervisor/types"
)

var _ InteropBackend = (*testutils.MockInteropBackend)(nil)

func TestInteropDeriver(t *testing.T) {
	logger := testlog.Logger(t, log.LevelInfo)
	l2Source := &testutils.MockL2Client{}
	emitter := &testutils.MockEmitter{}
	interopBackend := &testutils.MockInteropBackend{}
	cfg := &rollup.Config{
		InteropTime: new(uint64),
		L2ChainID:   big.NewInt(42),
	}
	chainID := supervisortypes.ChainIDFromBig(cfg.L2ChainID)

	rng := rand.New(rand.NewSource(123))
	genesisL1 := testutils.RandomBlockRef(rng)
	genesisL2 := testutils.RandomL2BlockRef(rng)
	anchor := AnchorPoint{
		CrossSafe:   genesisL2.BlockRef(),
		DerivedFrom: genesisL1,
	}
	loadAnchor := AnchorPointFn(func(ctx context.Context) (AnchorPoint, error) {
		return anchor, nil
	})
	interopDeriver := NewInteropDeriver(logger, cfg, context.Background(), interopBackend, l2Source, loadAnchor)
	interopDeriver.AttachEmitter(emitter)

	t.Run("local-unsafe blocks push to supervisor and trigger cross-unsafe attempts", func(t *testing.T) {
		emitter.ExpectOnce(engine.RequestCrossUnsafeEvent{})
		unsafeHead := testutils.NextRandomL2Ref(rng, 2, genesisL2, genesisL2.L1Origin)
		interopBackend.ExpectUpdateLocalUnsafe(chainID, unsafeHead.BlockRef(), nil)
		interopDeriver.OnEvent(engine.UnsafeUpdateEvent{Ref: unsafeHead})
		emitter.AssertExpectations(t)
		interopBackend.AssertExpectations(t)
	})
	t.Run("establish cross-unsafe", func(t *testing.T) {
		oldCrossUnsafe := testutils.NextRandomL2Ref(rng, 2, genesisL2, genesisL2.L1Origin)
		nextCrossUnsafe := testutils.NextRandomL2Ref(rng, 2, oldCrossUnsafe, oldCrossUnsafe.L1Origin)
		lastLocalUnsafe := testutils.NextRandomL2Ref(rng, 2, nextCrossUnsafe, nextCrossUnsafe.L1Origin)
		localView := supervisortypes.ReferenceView{
			Local: lastLocalUnsafe.ID(),
			Cross: oldCrossUnsafe.ID(),
		}
		supervisorView := supervisortypes.ReferenceView{
			Local: lastLocalUnsafe.ID(),
			Cross: nextCrossUnsafe.ID(),
		}
		interopBackend.ExpectUnsafeView(
			chainID, localView, supervisorView, nil)
		l2Source.ExpectL2BlockRefByHash(nextCrossUnsafe.Hash, nextCrossUnsafe, nil)
		emitter.ExpectOnce(engine.PromoteCrossUnsafeEvent{
			Ref: nextCrossUnsafe,
		})
		interopDeriver.OnEvent(engine.CrossUnsafeUpdateEvent{
			CrossUnsafe: oldCrossUnsafe,
			LocalUnsafe: lastLocalUnsafe,
		})
		interopBackend.AssertExpectations(t)
		emitter.AssertExpectations(t)
		l2Source.AssertExpectations(t)
	})
	t.Run("deny cross-unsafe", func(t *testing.T) {
		oldCrossUnsafe := testutils.NextRandomL2Ref(rng, 2, genesisL2, genesisL2.L1Origin)
		nextCrossUnsafe := testutils.NextRandomL2Ref(rng, 2, oldCrossUnsafe, oldCrossUnsafe.L1Origin)
		lastLocalUnsafe := testutils.NextRandomL2Ref(rng, 2, nextCrossUnsafe, nextCrossUnsafe.L1Origin)
		localView := supervisortypes.ReferenceView{
			Local: lastLocalUnsafe.ID(),
			Cross: oldCrossUnsafe.ID(),
		}
		supervisorView := supervisortypes.ReferenceView{
			Local: lastLocalUnsafe.ID(),
			Cross: oldCrossUnsafe.ID(), // stuck on same cross-safe
		}
		interopBackend.ExpectUnsafeView(
			chainID, localView, supervisorView, nil)
		interopDeriver.OnEvent(engine.CrossUnsafeUpdateEvent{
			CrossUnsafe: oldCrossUnsafe,
			LocalUnsafe: lastLocalUnsafe,
		})
		interopBackend.AssertExpectations(t)
		emitter.AssertExpectations(t) // no promote-cross-unsafe event expected
		l2Source.AssertExpectations(t)
	})
	t.Run("local-safe blocks push to supervisor and trigger cross-safe attempts", func(t *testing.T) {
		emitter.ExpectOnce(engine.RequestCrossSafeEvent{})
		derivedFrom := testutils.NextRandomRef(rng, genesisL1)
		localSafe := testutils.NextRandomL2Ref(rng, 2, genesisL2, genesisL2.L1Origin)
		interopBackend.ExpectUpdateLocalSafe(chainID, derivedFrom, localSafe.BlockRef(), nil)
		interopDeriver.OnEvent(engine.InteropPendingSafeChangedEvent{
			Ref:         localSafe,
			DerivedFrom: derivedFrom,
		})
		emitter.AssertExpectations(t)
		interopBackend.AssertExpectations(t)
	})
	t.Run("initialize cross-safe", func(t *testing.T) {
		oldCrossSafe := testutils.NextRandomL2Ref(rng, 2, genesisL2, genesisL2.L1Origin)
		nextCrossSafe := testutils.NextRandomL2Ref(rng, 2, oldCrossSafe, oldCrossSafe.L1Origin)
		lastLocalSafe := testutils.NextRandomL2Ref(rng, 2, nextCrossSafe, nextCrossSafe.L1Origin)
		localView := supervisortypes.ReferenceView{
			Local: lastLocalSafe.ID(),
			Cross: oldCrossSafe.ID(),
		}
		supervisorView := supervisortypes.ReferenceView{}
		interopBackend.ExpectSafeView(chainID, localView, supervisorView, supervisortypes.ErrUninitializedCrossSafeErr)
		interopBackend.ExpectInitializeCrossSafe(chainID, anchor.DerivedFrom, anchor.CrossSafe, nil)
		interopDeriver.OnEvent(engine.CrossSafeUpdateEvent{
			CrossSafe: oldCrossSafe,
			LocalSafe: lastLocalSafe,
		})
		interopBackend.AssertExpectations(t)
		emitter.AssertExpectations(t)
		l2Source.AssertExpectations(t)
	})
	t.Run("establish cross-safe", func(t *testing.T) {
		derivedFrom := testutils.NextRandomRef(rng, genesisL1)
		oldCrossSafe := testutils.NextRandomL2Ref(rng, 2, genesisL2, genesisL2.L1Origin)
		nextCrossSafe := testutils.NextRandomL2Ref(rng, 2, oldCrossSafe, oldCrossSafe.L1Origin)
		lastLocalSafe := testutils.NextRandomL2Ref(rng, 2, nextCrossSafe, nextCrossSafe.L1Origin)
		localView := supervisortypes.ReferenceView{
			Local: lastLocalSafe.ID(),
			Cross: oldCrossSafe.ID(),
		}
		supervisorView := supervisortypes.ReferenceView{
			Local: lastLocalSafe.ID(),
			Cross: nextCrossSafe.ID(),
		}
		interopBackend.ExpectSafeView(chainID, localView, supervisorView, nil)
		derived := eth.BlockID{
			Hash:   nextCrossSafe.Hash,
			Number: nextCrossSafe.Number,
		}
		interopBackend.ExpectDerivedFrom(chainID, derived, derivedFrom, nil)
		l2Source.ExpectL2BlockRefByHash(nextCrossSafe.Hash, nextCrossSafe, nil)
		emitter.ExpectOnce(engine.PromoteSafeEvent{
			Ref:         nextCrossSafe,
			DerivedFrom: derivedFrom,
		})
		emitter.ExpectOnce(engine.RequestFinalizedUpdateEvent{})
		interopDeriver.OnEvent(engine.CrossSafeUpdateEvent{
			CrossSafe: oldCrossSafe,
			LocalSafe: lastLocalSafe,
		})
		interopBackend.AssertExpectations(t)
		emitter.AssertExpectations(t)
		l2Source.AssertExpectations(t)
	})
	t.Run("deny cross-safe", func(t *testing.T) {
		oldCrossSafe := testutils.NextRandomL2Ref(rng, 2, genesisL2, genesisL2.L1Origin)
		nextCrossSafe := testutils.NextRandomL2Ref(rng, 2, oldCrossSafe, oldCrossSafe.L1Origin)
		lastLocalSafe := testutils.NextRandomL2Ref(rng, 2, nextCrossSafe, nextCrossSafe.L1Origin)
		localView := supervisortypes.ReferenceView{
			Local: lastLocalSafe.ID(),
			Cross: oldCrossSafe.ID(),
		}
		supervisorView := supervisortypes.ReferenceView{
			Local: lastLocalSafe.ID(),
			Cross: oldCrossSafe.ID(), // stay on old cross-safe
		}
		interopBackend.ExpectSafeView(chainID, localView, supervisorView, nil)
		interopDeriver.OnEvent(engine.CrossSafeUpdateEvent{
			CrossSafe: oldCrossSafe,
			LocalSafe: lastLocalSafe,
		})
		interopBackend.AssertExpectations(t)
		emitter.AssertExpectations(t) // no promote-cross-safe event expected
		l2Source.AssertExpectations(t)
	})
	t.Run("finalized L1 trigger cross-L2 finality check", func(t *testing.T) {
		emitter.ExpectOnce(engine.RequestFinalizedUpdateEvent{})
		finalizedL1 := testutils.RandomBlockRef(rng)
		interopBackend.ExpectUpdateFinalizedL1(chainID, finalizedL1, nil)
		interopDeriver.OnEvent(finality.FinalizeL1Event{
			FinalizedL1: finalizedL1,
		})
		emitter.AssertExpectations(t)
		interopBackend.AssertExpectations(t)
	})
	t.Run("next L2 finalized block", func(t *testing.T) {
		oldFinalizedL2 := testutils.NextRandomL2Ref(rng, 2, genesisL2, genesisL2.L1Origin)
		intermediateL2 := testutils.NextRandomL2Ref(rng, 2, oldFinalizedL2, oldFinalizedL2.L1Origin)
		nextFinalizedL2 := testutils.NextRandomL2Ref(rng, 2, intermediateL2, intermediateL2.L1Origin)
		emitter.ExpectOnce(engine.PromoteFinalizedEvent{
			Ref: nextFinalizedL2,
		})
		interopBackend.ExpectFinalized(chainID, nextFinalizedL2.ID(), nil)
		l2Source.ExpectL2BlockRefByHash(nextFinalizedL2.Hash, nextFinalizedL2, nil)
		interopDeriver.OnEvent(engine.FinalizedUpdateEvent{Ref: oldFinalizedL2})
		emitter.AssertExpectations(t)
		interopBackend.AssertExpectations(t)
	})
	t.Run("keep L2 finalized block", func(t *testing.T) {
		oldFinalizedL2 := testutils.NextRandomL2Ref(rng, 2, genesisL2, genesisL2.L1Origin)
		interopBackend.ExpectFinalized(chainID, oldFinalizedL2.ID(), nil)
		interopDeriver.OnEvent(engine.FinalizedUpdateEvent{Ref: oldFinalizedL2})
		emitter.AssertExpectations(t) // no PromoteFinalizedEvent
		interopBackend.AssertExpectations(t)
	})
}
