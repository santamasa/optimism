#!/bin/bash
set -eu

TOOL_NAME="xelper"
TOOL_VERSION="0.1.0"
CONTEXT_PREFIX="XELPER_CONTEXT"

# Arrays for storing command names and descriptions
COMMAND_NAMES=()
COMMAND_DESCRIPTIONS=()
register_command() {
    COMMAND_NAMES+=("$1")
    COMMAND_DESCRIPTIONS+=("$2")
}
# Load each command file
# Commands are expected to call register_command with the command name and description
for file in "$(dirname "$0")"/commands/*.sh; do
    source "$file"
done

print_help() {
    echo "$TOOL_NAME $TOOL_VERSION"
    echo "Usage: --context <context file> --print_context $TOOL_NAME <command> [options]"
    echo
    echo "Commands:"
    for i in "${!COMMAND_NAMES[@]}"; do
        printf "  %-10s %s\n" "${COMMAND_NAMES[$i]}" "${COMMAND_DESCRIPTIONS[$i]}"
    done
}

# Source standard setup
source "./setup.sh"
# Source the context file if provided
context_file=""
if [[ "$1" == "--context" ]]; then
  shift
  context_file="$1"
  if [[ -f "$context_file" ]]; then
    source "$context_file"
  fi
  shift
fi
echo $@
# Print Context if requested
if [[ "$1" == "--print_context" ]]; then
  printContext
  shift
fi

# Check if a command was given
if [[ $# -lt 1 ]]; then
    echo "Error: No command provided."
    print_help
    exit 1
fi

# move to the root directory of the monorepo
# this should be handled better later
pushd ../..
# Route commands dynamically
COMMAND="$1"
shift  # Shift arguments to access remaining ones for the command

# Search for the command in the COMMAND_NAMES array
command_found=0
for i in "${!COMMAND_NAMES[@]}"; do
    if [[ "${COMMAND_NAMES[$i]}" == "$COMMAND" ]]; then
        command_found=1
        output=$("$COMMAND" $@)
        break
    fi
done
if [[ $command_found -eq 0 ]]; then
    echo "Error: Unknown command '$COMMAND'"
    print_help
    popd
    exit 1
fi
popd

echo $output
if [[ -n "${context_file:-}" ]]; then
  saveContext "$context_file" "$output"
fi
