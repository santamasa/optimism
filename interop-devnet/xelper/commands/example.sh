# Register the 'example' command with its description
# register_command "example" <n1> <n2> - This is example's help message."

# Define the 'example' command function
example() {
  local context_variables=()
  if [[ $# -lt 2 ]]; then
      echo "Error: Missing arguments. Provide two numbers."
      echo "Usage: $TOOL_NAME sum <num1> <num2>"
      return 1
  fi

  local num1="$1"
  local num2="$2"
  local result=$((num1 + num2))
  # save the result as a context variable
  context_variables+=("result"="$result")
  # emit the context variables
  for var in "${context_variables[@]}"; do
      echo "$var"
  done
}
