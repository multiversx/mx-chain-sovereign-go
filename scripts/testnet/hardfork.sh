#!/usr/bin/env bash

export MULTIVERSXTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "$MULTIVERSXTESTNETSCRIPTSDIR/variables.sh"
source "$MULTIVERSXTESTNETSCRIPTSDIR/include/config.sh"

VALIDATOR_RES_PORT="$PORT_ORIGIN_VALIDATOR_REST"

if [ -z "$1" ]; then
  echo "epoch argument was not provided. Usage: './hardfork.sh [epoch number] [withEarlyEndOfEpoch: true|false]' as in './hardfork.sh 1 false'"
  exit
fi

if [ -z "$2" ]; then
  echo "forced argument was not provided. Usage: './hardfork.sh [epoch: number] [withEarlyEndOfEpoch: true|false]' as in './hardfork.sh 1 false'"
  exit
fi

if [ $1 -lt "1" ]; then
  echo "incorrect epoch argument was provided. Usage: './hardfork.sh [epoch number] [withEarlyEndOfEpoch: true|false]' as in './hardfork.sh 1 false'"
  exit
fi

address="http://127.0.0.1:$VALIDATOR_RES_PORT/hardfork/trigger"
epoch=$1
forced=$2
curl -d '{"epoch":'"$epoch"',"withEarlyEndOfEpoch":'"$forced"'}' -H 'Content-Type: application/json' $address

echo " done curl"

# change the setting from config.toml: AfterHardFork to true
updateTOMLValue "$TESTNETDIR/node/config/config_validator.toml" "AfterHardFork" "true"
updateTOMLValue "$TESTNETDIR/node/config/config_observer.toml" "AfterHardFork" "true"

# change nodesSetup.json genesis time to a new value
if [ "$ROUND_DURATION_IN_MS" -lt 1000 ]; then
  currentTimeMs=$(date +%s%3N)
  let "startTime = currentTimeMs + GENESIS_DELAY * 1000"
else
  currentTimeS=$(date +%s)
  let "startTime = currentTimeS + GENESIS_DELAY"
fi
updateJSONValue "$TESTNETDIR/node/config/nodesSetup.json" "startTime" "$startTime"

updateTOMLValue "$TESTNETDIR/node/config/config_validator.toml" "GenesisTime" $startTime
updateTOMLValue "$TESTNETDIR/node/config/config_observer.toml" "GenesisTime" $startTime

# copy back the configs
if [ $COPY_BACK_CONFIGS -eq 1 ]; then
  copyBackConfigs
fi

echo "done hardfork reconfig"
