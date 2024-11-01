# Xelper
Xelper is a scrappy bash framework for playing with the `interop-devnet`. The purpose is to allow interaction with deployed chains using composible scripting, while also supporting persisting the outputs in a way that's friendly for manual intervetion.

## Commands
Each Xelper command can be found in the `commands` directory. Commands are expected to adhere to the following
- Commands must `register_command` with their name and a help blurb. The name must match their entrypoint function.
- Commands are assumed to be running from the monorepo root directory.
- Commands may emit Context Variables.
- Commands may not use `echo` except to emit Context Variables.

## Context Variables
Xelper uses `context` files, which are glorified K:V stores of bash variables. When a command runs, it may
output Context Variables like so

```sh
context_variables+=("$chain"_EMITTER_ADDRESS="$address")
# emit the context variables
for var in "${context_variables[@]}"; do
    echo "$var"
done
```

This snippet is from the `deploy_emitter` command. Xelper will take these new keys, add some prefixing, and will save them to an ongoing context file.
Other commands can access these values with a `getContext` helper function that gets the stored value, like so:

```sh
  local emitter=$(getContext "$chain" EMITTER_ADDRESS)
  emitted=$(cast send --json --private-key=$RAW_PRIVATE_KEY $emitter 'emitData(bytes)' "$message")

```

This snippet is from the `emit` command, which will find a previously executed `deploy_emitter` output "EMITTER_ADDRESS" and can use it.

There are a few reasons for this context system:
- Allows for composible scripts to run with manual breaks or pauses.
- Allows for auditing of commands and their outputs over time for debugging.
- Allows for manual interaction with automatically created values, either by using the exported values manually, or by modifying the context values between executions.

When Xelper finishes execution, the new Context Variables are appended to the Context file. This means that historical values are preserved even when they would be overwritten by updates. Whenever context is passed into Xelper, it is imported by calling `source <context_file>`.

`--print_context` can be passed to Xelper to print the context values once sourced. This can be useful for seeing the current environment, or for cleaning up context files potentially.

## Example
For an example of how to use Xelper for complex workflows, check out `example_orchestration` which:
- Creates a new context,
- Establishes cross-dependency between chains,
- Checks each chain for liveness,
- Deploys the Emitter contract on one chain,
- Uses the Emitter contract on that chain,
- Creates an Executing Message on the other chain.

The context which is created can be used for further calls to Xelper, or you can manually use the outputs in your own work.

## The Future
I hit the edges of clean and manageable bash scripting almost immediately while building this. We should probably rewrite it in python or something.
