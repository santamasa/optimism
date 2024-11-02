import os
import json
import importlib
import argparse
from util import *


def list_commands():
  commands_dir = os.path.join(os.path.dirname(__file__), "commands")
  commands = []

  for filename in os.listdir(commands_dir):
    if filename.endswith(".py") and filename != "__init__.py":
      command_name = filename[:-3]
      module = importlib.import_module(f"commands.{command_name}")
      description = getattr(module, "description", "No description available.")
      commands.append((command_name, description))

  return commands

def run_command(context, subcmd):
  command = subcmd[0]
  args = subcmd[1:]
  try:
    module = importlib.import_module(f"commands.{command}")
    if hasattr(module, "execute"):
      return module.execute(context, args)
    else:
      print(f"Command '{command}' does not have an 'execute' function.")
  except ImportError:
    print(f"Command '{command}' not found.")

if __name__ == "__main__":
  parser = argparse.ArgumentParser(description="Xelper - Run Commands ; Persist Data")
  parser.add_argument("--context", type=str, help="location of context file", default="context.json")
  parser.add_argument("--verbose", action="store_true", default=False, help="show the output of the commands")
  parser.add_argument("subcmd", nargs=argparse.REMAINDER, help="Subcommand and its arguments")
  args = parser.parse_args()

  if not args.subcmd:
    print("Please provide a subcommand.")
    print("Available commands:")
    for command, description in list_commands():
      print(f"  {command}: {description}")
    exit(1)

  context = Context(args.context)
  (out, err) = run_command(context, args.subcmd)
  context.save()
  if args.verbose == True:
    print(f"{args.subcmd} - OUT:")
    print(out)
    print("--")
    print(f"{args.subcmd} - ERR:")
    print(err)
    print("--")

