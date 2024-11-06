package bootstrap

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"strings"

	artifacts2 "github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/artifacts"

	"github.com/ethereum-optimism/optimism/op-deployer/pkg/env"

	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/broadcaster"

	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer"
	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/opcm"
	opcrypto "github.com/ethereum-optimism/optimism/op-service/crypto"
	"github.com/ethereum-optimism/optimism/op-service/ctxinterrupt"
	"github.com/ethereum-optimism/optimism/op-service/ioutil"
	"github.com/ethereum-optimism/optimism/op-service/jsonutil"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"
)

type DisputeGameConfig struct {
	L1RPCUrl         string
	PrivateKey       string
	Logger           log.Logger
	ArtifactsLocator *artifacts2.Locator

	privateKeyECDSA *ecdsa.PrivateKey

	FPVM                     common.Address
	GameKind                 string
	GameType                 uint32
	AbsolutePrestate         common.Hash
	MaxGameDepth             uint64
	SplitDepth               uint64
	ClockExtension           uint64
	MaxClockDuration         uint64
	DelayedWethProxy         common.Address
	AnchorStateRegistryProxy common.Address
	L2ChainId                uint64
	Proposer                 common.Address
	Challenger               common.Address
}

func (c *DisputeGameConfig) Check() error {
	if c.L1RPCUrl == "" {
		return fmt.Errorf("l1RPCUrl must be specified")
	}

	if c.PrivateKey == "" {
		return fmt.Errorf("private key must be specified")
	}

	privECDSA, err := crypto.HexToECDSA(strings.TrimPrefix(c.PrivateKey, "0x"))
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}
	c.privateKeyECDSA = privECDSA

	if c.Logger == nil {
		return fmt.Errorf("logger must be specified")
	}

	if c.ArtifactsLocator == nil {
		return fmt.Errorf("artifacts locator must be specified")
	}

	if c.FPVM == (common.Address{}) {
		return fmt.Errorf("VM must be specified")
	}
	return nil
}

func DisputeGameCLI(cliCtx *cli.Context) error {
	logCfg := oplog.ReadCLIConfig(cliCtx)
	l := oplog.NewLogger(oplog.AppOut(cliCtx), logCfg)
	oplog.SetGlobalLogHandler(l.Handler())

	l1RPCUrl := cliCtx.String(deployer.L1RPCURLFlagName)
	privateKey := cliCtx.String(deployer.PrivateKeyFlagName)
	artifactsURLStr := cliCtx.String(ArtifactsLocatorFlagName)
	artifactsLocator := new(artifacts2.Locator)
	if err := artifactsLocator.UnmarshalText([]byte(artifactsURLStr)); err != nil {
		return fmt.Errorf("failed to parse artifacts URL: %w", err)
	}

	vm := common.HexToAddress(cliCtx.String(VMFlagName))
	if vm == (common.Address{}) {
		return fmt.Errorf("VM must be specified")
	}
	ctx := ctxinterrupt.WithCancelOnInterrupt(cliCtx.Context)

	return DisputeGame(ctx, DisputeGameConfig{
		L1RPCUrl:                 l1RPCUrl,
		PrivateKey:               privateKey,
		Logger:                   l,
		ArtifactsLocator:         artifactsLocator,
		FPVM:                     vm,
		GameKind:                 cliCtx.String(GameKindFlagName),
		GameType:                 uint32(cliCtx.Uint64(GameTypeFlagName)),
		AbsolutePrestate:         common.HexToHash(cliCtx.String(AbsolutePrestateFlagName)),
		MaxGameDepth:             cliCtx.Uint64(MaxGameDepthFlagName),
		SplitDepth:               cliCtx.Uint64(SplitDepthFlagName),
		ClockExtension:           cliCtx.Uint64(ClockExtensionFlagName),
		MaxClockDuration:         cliCtx.Uint64(MaxClockDurationFlagName),
		DelayedWethProxy:         common.HexToAddress(cliCtx.String(DelayedWethProxyFlagName)),
		AnchorStateRegistryProxy: common.HexToAddress(cliCtx.String(AnchorStateRegistryProxyFlagName)),
		L2ChainId:                cliCtx.Uint64(L2ChainIdFlagName),
		Proposer:                 common.HexToAddress(cliCtx.String(ProposerFlagName)),
		Challenger:               common.HexToAddress(cliCtx.String(ChallengerFlagName)),
	})
}

func DisputeGame(ctx context.Context, cfg DisputeGameConfig) error {
	if err := cfg.Check(); err != nil {
		return fmt.Errorf("invalid config for DisputeGame: %w", err)
	}

	lgr := cfg.Logger
	progressor := func(curr, total int64) {
		lgr.Info("artifacts download progress", "current", curr, "total", total)
	}

	artifactsFS, cleanup, err := artifacts2.Download(ctx, cfg.ArtifactsLocator, progressor)
	if err != nil {
		return fmt.Errorf("failed to download artifacts: %w", err)
	}
	defer func() {
		if err := cleanup(); err != nil {
			lgr.Warn("failed to clean up artifacts", "err", err)
		}
	}()

	l1Client, err := ethclient.Dial(cfg.L1RPCUrl)
	if err != nil {
		return fmt.Errorf("failed to connect to L1 RPC: %w", err)
	}

	chainID, err := l1Client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %w", err)
	}

	signer := opcrypto.SignerFnFromBind(opcrypto.PrivateKeySignerFn(cfg.privateKeyECDSA, chainID))
	chainDeployer := crypto.PubkeyToAddress(cfg.privateKeyECDSA.PublicKey)

	bcaster, err := broadcaster.NewKeyedBroadcaster(broadcaster.KeyedBroadcasterOpts{
		Logger:  lgr,
		ChainID: chainID,
		Client:  l1Client,
		Signer:  signer,
		From:    chainDeployer,
	})
	if err != nil {
		return fmt.Errorf("failed to create broadcaster: %w", err)
	}

	nonce, err := l1Client.NonceAt(ctx, chainDeployer, nil)
	if err != nil {
		return fmt.Errorf("failed to get starting nonce: %w", err)
	}

	host, err := env.DefaultScriptHost(
		bcaster,
		lgr,
		chainDeployer,
		artifactsFS,
		nonce,
	)
	if err != nil {
		return fmt.Errorf("failed to create script host: %w", err)
	}

	dgo, err := opcm.DeployDisputeGame(
		host,
		opcm.DeployDisputeGameInput{
			FpVm:                     cfg.FPVM,
			GameKind:                 cfg.GameKind,
			GameType:                 cfg.GameType,
			AbsolutePrestate:         cfg.AbsolutePrestate,
			MaxGameDepth:             cfg.MaxGameDepth,
			SplitDepth:               cfg.SplitDepth,
			ClockExtension:           cfg.ClockExtension,
			MaxClockDuration:         cfg.MaxClockDuration,
			DelayedWethProxy:         cfg.DelayedWethProxy,
			AnchorStateRegistryProxy: cfg.AnchorStateRegistryProxy,
			L2ChainId:                cfg.L2ChainId,
			Proposer:                 cfg.Proposer,
			Challenger:               cfg.Challenger,
		},
	)
	if err != nil {
		return fmt.Errorf("error deploying dispute game: %w", err)
	}

	if _, err := bcaster.Broadcast(ctx); err != nil {
		return fmt.Errorf("failed to broadcast: %w", err)
	}

	lgr.Info("deployed dispute game")

	if err := jsonutil.WriteJSON(dgo, ioutil.ToStdOut()); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}
	return nil
}
