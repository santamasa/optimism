register_command "deploy_emitter" "<chainID> - Puts the Emitter Contract on the specified chain"

deploy_emitter() {
  local context_variables=()
  local chain="$1"

  export OP_INTEROP_DEVKEY_CHAINID=0
  export OP_INTEROP_DEVKEY_DOMAIN=user
  export OP_INTEROP_DEVKEY_NAME=0
  export ETH_RPC_URL=$(urlForChain "$chain")
  export RAW_PRIVATE_KEY=0x$(go run ./op-node/cmd interop devkey secret)

  # deploy the data emitter contract to chain A
  pushd op-e2e/e2eutils/interop/contracts/
  # consider putting forge build somewhere else so it can be reused
  build=$(forge build)
  create=$(forge create --private-key=$RAW_PRIVATE_KEY "src/emit.sol:EmitEvent" --json)
  popd

  # save the address of the deployed contract
  address=$(echo "$create" | jq -r .deployedTo)

  context_variables+=("$chain"_EMITTER_CREATE="$create")
  context_variables+=("$chain"_EMITTER_ADDRESS="$address")
  # emit the context variables
  for var in "${context_variables[@]}"; do
      echo "$var"
  done
}
