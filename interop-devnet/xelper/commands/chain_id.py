import subprocess
import argparse
import json
from util import RPC_URLs, MONOREPO_ROOT

description = "Get the chain ID of a chain from cast."

def execute(context, args):
  parser = argparse.ArgumentParser()
  parser.add_argument('chain', type=str, help="Chain to deploy the emitter contract to")
  args = parser.parse_args(args)

  rpc_url = RPC_URLs[args.chain]
  root = MONOREPO_ROOT

  cmd = f"""
  #!/bin/bash
  cd {root}
  OP_INTEROP_DEVKEY_CHAINID=0
  OP_INTEROP_DEVKEY_DOMAIN=user
  OP_INTEROP_DEVKEY_NAME=0
  RAW_PRIVATE_KEY=0x$(go run ./op-node/cmd interop devkey secret)
  ETH_RPC_URL={rpc_url}
  cast chain-id
  """

  result = subprocess.run(cmd, capture_output=True, executable='/bin/bash', shell=True, text=True)

  context.set(f"{args.chain}.ChainID", result.stdout.strip())

  print(f"Chain ID: {result.stdout.strip()}")
  return(result.stdout, result.stderr)
