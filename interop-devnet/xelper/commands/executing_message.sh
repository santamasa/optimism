register_command "executing_message" "<destination chainID> <source chainID> <key> - Creates an Executing Message on the specified chain for Log Event"

executing_message() {
  local context_variables=()
  local chain="$1"
  local srcChain="$2"
  local key="$3"

  export OP_INTEROP_DEVKEY_CHAINID=$chain
  export OP_INTEROP_DEVKEY_DOMAIN=user
  export OP_INTEROP_DEVKEY_NAME=0
  export ETH_RPC_URL=$(urlForChain "$chain")
  export RAW_PRIVATE_KEY=0x$(go run ./op-node/cmd interop devkey secret)

  local logEvent=$(getContext "$srcChain" EMITTED "$key")

  # TODO - these jq calls aren't all set up correctly yet
  local blockNum=$(echo logEvent | jq -r .blockNumber)
  local index=$(echo logEvent | jq .)
  local timestamp=$(cast block -f timestamp $blockNum)
  topic1 = $(echo logEvent | jq -r .topics[0])
  topic2 = $(echo logEvent | jq -r .topics[1])
  combinedHash=$(cast concat-hex "$topic1" "$topic2")
  address=$(echo logEvent | jq -r .address)

  executing=$(cast send --json --private-key=$RAW_PRIVATE_KEY $CROSS_L2_INBOX_ADDRESS \
  executeMessage((address,uint256,uint256,uint256,uint256),address,bytes) \
    ("$address", "$blockNum", "$index", "$timestamp", "$srcChain") \
    0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa \
    "$combinedHash")

  context_variables+=("$chain"_EXECUTING_"$key"="$executing")
  # emit the context variables
  for var in "${context_variables[@]}"; do
      echo "$var"
  done
}
