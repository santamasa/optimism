import subprocess

description = "List the contents of a directory and save to context."

def execute(context, args):
    print(args)
    # Run the 'ls' command and capture output
    result = subprocess.run(["ls"], capture_output=True, text=True)

    # Save the output to the context with key 'directory_listing'
    context.set("directory_listing", result.stdout.splitlines())
    print("Directory listing saved to context.")
