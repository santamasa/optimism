register_command "check_chain" "<chainID> - Accesses a chain and prints basic information about it"

check_chain() {
  local context_variables=()
  local chain="$1"

  export OP_INTEROP_DEVKEY_CHAINID=0
  export OP_INTEROP_DEVKEY_DOMAIN=user
  export OP_INTEROP_DEVKEY_NAME=0
  export RAW_PRIVATE_KEY=0x$(go run ./op-node/cmd interop devkey secret)

  export ETH_RPC_URL=$(urlForChain "$chain")
  local chainID=$(cast chain-id)
  context_variables+=("$chain"_CHAINID="$chainID")

  # emit the context variables
  for var in "${context_variables[@]}"; do
      echo "$var"
  done

  export new_context="BBBBBB"
}
