#!/bin/bash

context="$(uuidgen).context"

./xelper.sh --context "$context" add_cross_dependency 900200 900201

./xelper.sh --context "$context" check_chain 900200

./xelper.sh --context "$context" check_chain 900201

./xelper.sh --context "$context" deploy_emitter 900200

./xelper.sh --context "$context" emit 900200 SAMPLE_LOG "Hello, World!"

./xelper.sh --context "$context" executing_message 900201 900200 SAMPLE_LOG
