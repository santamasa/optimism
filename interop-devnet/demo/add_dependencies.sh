#!/bin/bash

set -eu

export OP_INTEROP_DEVKEY_CHAINID=900200
export OP_INTEROP_DEVKEY_DOMAIN=chain-operator
export OP_INTEROP_DEVKEY_NAME=system-config-owner
export OP_INTEROP_MNEMONIC="test test test test test test test test test test test junk"
export ETH_RPC_URL="http://yolo:8545"
export RAW_PRIVATE_KEY=0x$(go run ./op-node/cmd interop devkey secret)
export SYSTEM_CONFIG_ADDR=$(jq -r .SystemConfigProxy .devnet-interop/deployments/l2/900200/addresses.json)
cast send --private-key=$RAW_PRIVATE_KEY $SYSTEM_CONFIG_ADDR 'addDependency(uint256)' 900200
cast send --private-key=$RAW_PRIVATE_KEY $SYSTEM_CONFIG_ADDR 'addDependency(uint256)' 900201


export OP_INTEROP_DEVKEY_CHAINID=900201
export RAW_PRIVATE_KEY=0x$(go run ./op-node/cmd interop devkey secret)
export SYSTEM_CONFIG_ADDR=$(jq -r .SystemConfigProxy .devnet-interop/deployments/l2/900201/addresses.json)
cast send --private-key=$RAW_PRIVATE_KEY $SYSTEM_CONFIG_ADDR 'addDependency(uint256)' 900200
cast send --private-key=$RAW_PRIVATE_KEY $SYSTEM_CONFIG_ADDR 'addDependency(uint256)' 900201
