#!/bin/bash

set -eu

export OP_INTEROP_DEVKEY_CHAINID=0
export OP_INTEROP_DEVKEY_DOMAIN=user
export OP_INTEROP_DEVKEY_NAME=0
export OP_INTEROP_MNEMONIC="test test test test test test test test test test test junk"
export RAW_PRIVATE_KEY=0x$(go run ./op-node/cmd interop devkey secret)

# deploy the data emitter contract to chain A
cd op-e2e/e2eutils/interop/contracts/
forge build
export ETH_RPC_URL="http://yolo:9145"
forge create --private-key=$RAW_PRIVATE_KEY "src/emit.sol:EmitEvent"
