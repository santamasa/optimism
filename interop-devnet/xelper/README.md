# Xelper
Xelper is a lightweight python/shell-scripting framework for playing with the `interop-devnet`. The purpose is to allow interaction with deployed chains using composable scripting, while also supporting persisting the outputs in a way that's friendly for manual intervetion.

## Commands
Each Xelper command can be found in the `commands` directory. Commands are expected to adhere to the following
- Commands are python modules which must satisfy `execute` and must contain a `description`
- Commands recieve commandline arguments, and a Context dictionary
- Commands may do whatever they wish during execution
- Commands may (and are encouraged to) save new values to the supplied Context

## Context Variables
Xelper uses `context` files, which is just a named JSON format K:V storage that commands can share. For example:

```sh
  forge create --private-key=$RAW_PRIVATE_KEY "src/emit.sol:EmitEvent" --json
```
```py
  result = subprocess.run(cmd, capture_output=True, executable='/bin/bash', shell=True, text=True)
  resultJSON = json.loads(result.stdout)

  context.set(f"{args.chain}.CreateEmitterOutput", resultJSON)
  context.set(f"{args.chain}.EmitterContractAddress", resultJSON["deployedTo"])
```
This snippet is from the `deploy_emitter` command. The command uses subprocess to run a `forge` contract creation, and then saves the result structure, *and* saves the contract to a separate key.

When the command is done, Xelper will save the new version of the Context to the original location. Other commands can access these new values to extend the previous steps, like so:

```py
  emitter = context.get(f"{args.chain}.EmitterContractAddress")
```
```sh
  cast send --json --private-key=$RAW_PRIVATE_KEY {emitter} 'emitData(bytes)' "$message"
```
This snippet is from `emit`, which first locates a previously registered Emitter contract, and then uses string-formatting to include it in a subprocess call to `cast`

All the existing commands follow the same pattern of leveraging subprocess to execute bash scripts, and then consume the output and persist important parts to the Context.

There are a few reasons for this context system:
- Script execution can be paused and extended without lengthy re-runs. Existing context can be used to seed or recover automation.
- In JSON format, the Context is human readable, and human Editable. If you wanted to influence a test but don't want to write a whole new workflow, Context files are hacker friendly.
- Context files can be copied and reused, allowing for longer running tests to fork the context and leverage shared resources.

## Running
```
python3 ./xelper [top level arguments] <subcommand> [subcommand arguments]
```
To see the full list of commands, run Xelper without specifying a subcommand.

### Optional Top Level Arguments
**--context** specifies a context file, or will default to `context.json`
**--verbose** will print the full `stdout` and `stderr` returned from subcommands

## Example
For an example of how to use Xelper for complex workflows, check out `example_orchestration.sh` which:
- Creates a new context,
- Establishes cross-dependency between chains,
- Checks each chain for liveness,
- Deploys the Emitter contract on one chain,
- Uses the Emitter contract on that chain,
- Creates an Executing Message on the other chain.

The context which is created can be used for further calls to Xelper, or you can manually use the outputs in your own work. You could also put manual pause steps between thee lines, or add branching based on the Context. The point of Xelper is to allow composable scripts to build off one another without making any monolithic scripts
