register_command "emit" "<chainID> <key> <message> - Creates a Log Event on the specified chain with the specified key"

emit() {
  local context_variables=()
  local chain="$1"
  local key="$2"
  local data="$@"

  export OP_INTEROP_DEVKEY_CHAINID=$chain
  export OP_INTEROP_DEVKEY_DOMAIN=user
  export OP_INTEROP_DEVKEY_NAME=0
  export ETH_RPC_URL=$(urlForChain "$chain")
  export RAW_PRIVATE_KEY=0x$(go run ./op-node/cmd interop devkey secret)

  local emitter=$(getContext "$chain" EMITTER_ADDRESS)

  message=$(cast from-utf8 "$data")

  emitted=$(cast send --json --private-key=$RAW_PRIVATE_KEY $emitter 'emitData(bytes)' "$message")

  context_variables+=("$chain"_EMITTED_"$key"="$emitted")
  # emit the context variables
  for var in "${context_variables[@]}"; do
      echo "$var"
  done
}
