import os
import json

# TODO this should become a config file
L1_ETH_RPC_URL="http://0.0.0.0:8545"
L2_A_RPC_URL="http://0.0.0.0:9145" # chain A
L2_B_RPC_URL="http://0.0.0.0:9245" # chain B

RPC_URLs = {
  # L1 RPC and Aliases
  "L1": L1_ETH_RPC_URL,

  # L2 Chain A RPC and Aliases
  "L2_A": L2_A_RPC_URL,
  "900200": L2_A_RPC_URL,

  # L2 Chain B RPC and Aliases
  "L2_B": L2_B_RPC_URL,
  "900201": L2_B_RPC_URL
}

class Context:
    def __init__(self, filepath="context.json"):
        self.filepath = filepath
        self.data = self._load_context()

    def _load_context(self):
        if os.path.exists(self.filepath):
            with open(self.filepath, "r") as f:
                return json.load(f)
        return {}

    def save(self):
        with open(self.filepath, "w") as f:
            json.dump(self.data, f, indent=4)

    def set(self, key, value):
        self.data[key] = value
        self.save()

    def get(self, key):
        return self.data.get(key)
