
# Set common environment variables
export OP_INTEROP_MNEMONIC="test test test test test test test test test test test junk"
# TODO: should gather this more dynamically
export L1_ETH_RPC_URL="http://0.0.0.0:8545"
export L2_A_RPC_URL="http://0.0.0.0:9145" # chain A
export L2_B_RPC_URL="http://0.0.0.0:9245" # chain B

# Make a lookup for chain IDs to L2 RPC URLs
ID_LIST=(900200 900201)
URL_LIST=($L2_A_RPC_URL $L2_B_RPC_URL)
urlForChain() {
    local id="$1"
    # Loop through ID_LIST to find the index of the given ID
    for i in "${!ID_LIST[@]}"; do
        if [[ "${ID_LIST[$i]}" == "$id" ]]; then
            echo "${URL_LIST[$i]}"
            return 0
        fi
    done
    # If ID not found, output an error
    echo "Error: No URL found for ID '$id'" >&2
    return 1
}

# returns the xelper context variable
# Usage: getContext 900200 CHAINID
# This will return the value of the variable XELPER_CONTEXT_900200_CHAINID
getContext() {
  local concat="$CONTEXT_PREFIX"
  for value in "$@"; do
      concat="$concat"_"$value"
  done
  value=$(eval echo \$$concat)
  echo "$value"
}

printContext() {
  for var in $(compgen -v | grep "^$CONTEXT_PREFIX"); do
    echo "$var=${!var}"
  done
}

saveContext() {
  local prefix="$CONTEXT_PREFIX"
  local filename="$1"
  local output="$2"
  if [[ -n "${output:-}" ]]; then
    echo "Saving context variables to '$context_file'..."
    timestamp=$(date +%s)
    echo "# new context variables: $timestamp" >> "$context_file"
    while IFS='=' read -r key value; do
        echo "$prefix"_"$key"="$value" >> "$context_file"
    done <<< "$output"
fi
}

# overload pushd and popd to suppress output
# TODO: not all environments support pushd/popd
# a better traversal method would be nice
pushd () {
    command pushd "$@" > /dev/null
}
popd () {
    command popd "$@" > /dev/null
}
export pushd popd
