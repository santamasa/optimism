import subprocess
import argparse
import json
from util import RPC_URLs, MONOREPO_ROOT, CROSS_L2_INBOX_ADDRESS

description = "Create an Executing Message referencing a prior Initiating Message."

def execute(context, args):
  parser = argparse.ArgumentParser()
  parser.add_argument('chainA', type=str, help="Chain A")
  parser.add_argument('chainB', type=str, help="Chain B")
  args = parser.parse_args(args)

  rpc_url = RPC_URLs["L1"]
  root = MONOREPO_ROOT
  A = args.chainA
  B = args.chainB

  cmd = f"""
  #!/bin/bash
  cd {root}
  OP_INTEROP_MNEMONIC="test test test test test test test test test test test junk"
  export OP_INTEROP_DEVKEY_DOMAIN=chain-operator
  export OP_INTEROP_DEVKEY_NAME=system-config-owner
  ETH_RPC_URL={rpc_url}

  export OP_INTEROP_DEVKEY_CHAINID={A}
  export RAW_PRIVATE_KEY=0x$(go run ./op-node/cmd interop devkey secret)
  export SYSTEM_CONFIG_ADDR=$(jq -r .SystemConfigProxy .devnet-interop/deployments/l2/{A}/addresses.json)
  a_a=$(cast send --private-key=$RAW_PRIVATE_KEY $SYSTEM_CONFIG_ADDR 'addDependency(uint256)' {A} --json)
  a_b=$(cast send --private-key=$RAW_PRIVATE_KEY $SYSTEM_CONFIG_ADDR 'addDependency(uint256)' {B} --json)

  export OP_INTEROP_DEVKEY_CHAINID={B}
  export RAW_PRIVATE_KEY=0x$(go run ./op-node/cmd interop devkey secret)
  export SYSTEM_CONFIG_ADDR=$(jq -r .SystemConfigProxy .devnet-interop/deployments/l2/{B}/addresses.json)
  cast send --private-key=$RAW_PRIVATE_KEY $SYSTEM_CONFIG_ADDR 'addDependency(uint256)' {A} --json
  cast send --private-key=$RAW_PRIVATE_KEY $SYSTEM_CONFIG_ADDR 'addDependency(uint256)' {B} --json
  """
  result = subprocess.run(cmd, capture_output=True, executable='/bin/bash', shell=True, text=True)

  # TODO: this is not a stable way to check if the command was successful
  if result.stderr == "":
    context.set(f"ChainDependency.{A}.{B}", True)
    context.set(f"ChainDependency.{B}.{A}", True)

  print(f"Superchain Dependencies Created ({A}:{B})")
  return(result.stdout, result.stderr)
