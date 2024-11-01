register_command "add_cross_dependency" "<chainID> <chainID> - Establishes cross-chain dependencies between two chains"

add_cross_dependency() {
  local context_variables=()
  local chainA="$1"
  local chainB="$2"

  export OP_INTEROP_DEVKEY_DOMAIN=chain-operator
  export OP_INTEROP_DEVKEY_NAME=system-config-owner

  # Sending to L1
  export ETH_RPC_URL=$L1_ETH_RPC_URL

  # Assume Chain 1 and send addDependency calls to the L1
  export OP_INTEROP_DEVKEY_CHAINID=$1
  export RAW_PRIVATE_KEY=0x$(go run ./op-node/cmd interop devkey secret)
  export SYSTEM_CONFIG_ADDR=$(jq -r .SystemConfigProxy .devnet-interop/deployments/l2/$1/addresses.json)
  a_a=$(cast send --private-key=$RAW_PRIVATE_KEY $SYSTEM_CONFIG_ADDR 'addDependency(uint256)' $1 --json)
  a_b=$(cast send --private-key=$RAW_PRIVATE_KEY $SYSTEM_CONFIG_ADDR 'addDependency(uint256)' $2 --json)

  # Assume Chain 2 and send addDependency calls to the L1
  export OP_INTEROP_DEVKEY_CHAINID=$2
  export RAW_PRIVATE_KEY=0x$(go run ./op-node/cmd interop devkey secret)
  export SYSTEM_CONFIG_ADDR=$(jq -r .SystemConfigProxy .devnet-interop/deployments/l2/$2/addresses.json)
  b_a=$(cast send --private-key=$RAW_PRIVATE_KEY $SYSTEM_CONFIG_ADDR 'addDependency(uint256)' $1 --json)
  b_b=$(cast send --private-key=$RAW_PRIVATE_KEY $SYSTEM_CONFIG_ADDR 'addDependency(uint256)' $2 --json)

  # save all addDependency results to context variables
  context_variables+=("$chainA"_"$1"_addDependency="$a_a")
  context_variables+=("$chainA"_"$2"_addDependency="$a_b")
  context_variables+=("$chainB"_"$1"_addDependency="$b_a")
  context_variables+=("$chainB"_"$2"_addDependency="$b_b")
  # emit the context variables
  for var in "${context_variables[@]}"; do
      echo "$var"
  done
}
