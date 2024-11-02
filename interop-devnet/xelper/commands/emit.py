import subprocess
import argparse
import json
from util import RPC_URLs, MONOREPO_ROOT

description = "Use the Emitter Contract to emit an event."

def execute(context, args):
  parser = argparse.ArgumentParser()
  parser.add_argument('chain', type=str, help="Chain to deploy the emitter contract to")
  parser.add_argument('--key', type=str, default="default", help="Idempotency key for event lookup")
  parser.add_argument('--data', type=str, default="Hello Superchain!", help="Data to emit")
  args = parser.parse_args(args)

  rpc_url = RPC_URLs[args.chain]
  root = MONOREPO_ROOT
  emitter = context.get(f"{args.chain}.EmitterContractAddress")
  if not emitter:
    print("Emitter contract not found in context. Please deploy the emitter contract first.")
    exit(1)

  cmd = f"""
  #!/bin/bash
  cd {root}
  OP_INTEROP_MNEMONIC="test test test test test test test test test test test junk"
  OP_INTEROP_DEVKEY_CHAINID={args.chain}
  OP_INTEROP_DEVKEY_DOMAIN=user
  OP_INTEROP_DEVKEY_NAME=0
  ETH_RPC_URL={rpc_url}
  RAW_PRIVATE_KEY=0x$(go run ./op-node/cmd interop devkey secret)

  message=$(cast from-utf8 "{args.data}")
  cast send --json --private-key=$RAW_PRIVATE_KEY {emitter} 'emitData(bytes)' "$message"
  """

  result = subprocess.run(cmd, capture_output=True, executable='/bin/bash', shell=True, text=True)
  resultJSON = json.loads(result.stdout)

  context.set(f"{args.chain}.Emitted.{args.key}", resultJSON)

  print(f"Event emitted: {resultJSON}")
  return(result.stdout, result.stderr)
