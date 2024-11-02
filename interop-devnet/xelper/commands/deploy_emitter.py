import subprocess
import argparse
import json
from util import RPC_URLs

description = "List the contents of a directory and save to context."

def execute(context, args):
  parser = argparse.ArgumentParser(description="deploy_emitter - Deploy the emitter contract")
  parser.add_argument('chain', type=str, help="Chain to deploy the emitter contract to")
  args = parser.parse_args(args)

  rpc_url = RPC_URLs[args.chain]

  cmd = f"""
  -eu
  cd ../..
  export OP_INTEROP_MNEMONIC="test test test test test test test test test test test junk"
  export OP_INTEROP_DEVKEY_CHAINID=0
  export OP_INTEROP_DEVKEY_DOMAIN=user
  export OP_INTEROP_DEVKEY_NAME=0
  export ETH_RPC_URL={rpc_url}
  export RAW_PRIVATE_KEY=0x$(go run ./op-node/cmd interop devkey secret)
  cd op-e2e/e2eutils/interop/contracts/
  # TODO: make forge build its own command, this takes too much time
  #forge build 2&>1 /dev/null
  forge create --private-key=$RAW_PRIVATE_KEY "src/emit.sol:EmitEvent" --json
  cd ../../../..
  """
  result = subprocess.run(cmd, capture_output=True, shell=True, text=True)
  print(result.stdout)

  resultJSON = json.loads(result.stdout)

  # Save the output to the context with key 'directory_listing'
  context.set(f"{args.chain}.CreateEmitterOutput", result.stdout)
  context.set(f"{args.chain}.EmitterContractAddress", resultJSON["deployedTo"])
  print("Directory listing saved to context.")
