#!/bin/bash

# Current location
SCRIPT_PATH="$(dirname "$(realpath "$BASH_SOURCE")")"

# Source node variables
TESTNET_DIR="$(dirname $SCRIPT_PATH)"
source "$TESTNET_DIR/variables.sh"
source "$TESTNET_DIR/include/config.sh"

# Source all scripts
source "$SCRIPT_PATH/config/configs.cfg"
source "$SCRIPT_PATH/config/helper.cfg"
source "$SCRIPT_PATH/config/utils.snippets.sh"
source "$SCRIPT_PATH/config/py.snippets.sh"
source "$SCRIPT_PATH/config/deploy.snippets.sh"
source "$SCRIPT_PATH/observer/deployObserver.sh"
source "$SCRIPT_PATH/config/sovereign.snippets.sh"

# Create necessary directories
mkdir -p "$SOVEREIGN_DIRECTORY"
mkdir -p "$OUTFILE_PATH"
mkdir -p "$CONTRACTS_DIRECTORY"

# Define other variables
WALLET_ADDRESS=$(echo "$(head -n 1 "$WALLET")" | sed -n 's/.* for \([^-]*\)-----.*/\1/p')
