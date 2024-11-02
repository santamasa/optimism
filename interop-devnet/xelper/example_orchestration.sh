#!/bin/bash

context="$(uuidgen).context.json"
#verbose="--verbose"

python3 ./xelper.py --context "$context" $verbose add_dependency 900200 900201

python3 ./xelper.py --context "$context" $verbose chain_id 900200

python3 ./xelper.py --context "$context" $verbose chain_id 900201

python3 ./xelper.py --context "$context" $verbose deploy_emitter 900200

python3 ./xelper.py --context "$context" $verbose emit 900200 --data "Hello, World!" --key EXAMPLE

python3 ./xelper.py --context "$context" $verbose executing_message 900201 900200 --key EXAMPLE

echo "Orchestration complete. Context file: $context"
cat "$context" | jq .
