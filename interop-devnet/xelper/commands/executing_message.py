import subprocess
import argparse
import json
from util import RPC_URLs, MONOREPO_ROOT, CROSS_L2_INBOX_ADDRESS

description = "Create an Executing Message referencing a prior Initiating Message."

def execute(context, args):
  parser = argparse.ArgumentParser()
  parser.add_argument('chain', type=str, help="Chain to put this Executing Message on")
  parser.add_argument('src_chain', type=str, help="Chain where the Initiating Message is")
  parser.add_argument('--key', type=str, default="default", help="Idempotency key for event lookup")
  args = parser.parse_args(args)

  rpc_url = RPC_URLs[args.chain]
  root = MONOREPO_ROOT
  iEventKey = f"{args.src_chain}.Emitted.{args.key}"
  iEvent = context.get(iEventKey)
  if not iEvent:
    print(f"Initiating Messge ({iEventKey}) not found in context.")
    exit(1)
  log = iEvent["logs"][0]
  blockNumber = int(iEvent["blockNumber"], 16)

  cmd = f"""
  #!/bin/bash
  cd {root}
  OP_INTEROP_MNEMONIC="test test test test test test test test test test test junk"
  OP_INTEROP_DEVKEY_CHAINID={args.chain}
  OP_INTEROP_DEVKEY_DOMAIN=user
  OP_INTEROP_DEVKEY_NAME=0
  ETH_RPC_URL={rpc_url}
  RAW_PRIVATE_KEY=0x$(go run ./op-node/cmd interop devkey secret)

  combined_hash=$(cast concat-hex {log["topics"][0]} {log["topics"][1]})
  timestamp=$(cast block -f timestamp {blockNumber})

  cast send --json --private-key=$RAW_PRIVATE_KEY {CROSS_L2_INBOX_ADDRESS} \
  "executeMessage((address,uint256,uint256,uint256,uint256),address,bytes)" \
    "({log["address"]}, {blockNumber}, 0, $timestamp, {args.src_chain})" \
    '0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa' \
    "$combined_hash"

  """
  result = subprocess.run(cmd, capture_output=True, executable='/bin/bash', shell=True, text=True)
  resultJSON = json.loads(result.stdout)

  context.set(f"{args.chain}.ExeMessage.{args.src_chain}.{args.key}", resultJSON)

  print(f"Executing Message created: {resultJSON}")
  return(result.stdout, result.stderr)
