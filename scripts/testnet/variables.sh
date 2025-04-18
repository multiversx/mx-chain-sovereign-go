# These paths must be absolute

# METASHARD_ID will be used to identify a shard ID as metachain
export METASHARD_ID=4294967295

# Path to mx-chain-go. Determined automatically. Do not change.
export MULTIVERSXDIR=$(dirname $(dirname $MULTIVERSXTESTNETSCRIPTSDIR))

# Enable the MultiversX Proxy. Note that this is a private repository
# (mx-chain-proxy-go).
export USE_PROXY=1

# Enable the MultiversX Transaction Generator. Note that this is a private
# repository (mx-chain-txgen-go).
export USE_TXGEN=0

# Enable the Elasticsearch data indexing. Will run a Docker image containing an Elasticsearch cluster, on port 9200.
# It will also change the external.toml files for observers, so they can index data into it.
# Docker must be managed as a non-root user: https://docs.docker.com/engine/install/linux-postinstall/
export USE_ELASTICSEARCH=0

# Elasticsearch volume name to keep the elastic history on host. History will be loaded when docker is starting.
export ELASTICSEARCH_VOLUME="sov-elastic"

# Path where the testnet will be instantiated. This folder is assumed to not
# exist, but it doesn't matter if it already does. It will be created if not,
# anyway.
export TESTNETDIR="$HOME/MultiversX/testnet"

# Path to mx-chain-deploy-go, branch: master. Default: near mx-chain-go.
export CONFIGGENERATORDIR="$(dirname $MULTIVERSXDIR)/mx-chain-deploy-go/cmd/filegen"
export CONFIGGENERATOR="$CONFIGGENERATORDIR/filegen"    # Leave unchanged.
export CONFIGGENERATOROUTPUTDIR="output"

# Path to the executable node. Leave unchanged unless well justified.
export NODEDIR="$MULTIVERSXDIR/cmd/node"
export NODE="$NODEDIR/node"     # Leave unchanged

# Path to the executable sovereign node. Leave unchanged unless well justified.
export SOVEREIGNNODEDIR="$MULTIVERSXDIR/cmd/sovereignnode"
export SOVEREIGNNODE="$SOVEREIGNNODEDIR/sovereignnode"     # Leave unchanged

# Path to the executable seednode. Leave unchanged unless well justified.
export SEEDNODEDIR="$MULTIVERSXDIR/cmd/seednode"
export SEEDNODE="$SEEDNODEDIR/seednode"   # Leave unchanged.

# Niceness value of the Seednode, Observer Nodes and Validator Nodes. Leave
# blank to not adjust niceness.
export NODE_NICENESS=10

# Start a watcher daemon for each validator node, which restarts the node if it
# is suffled out of its shard.
export NODE_WATCHER=0

# Delays after running executables.
export SEEDNODE_DELAY=5
export GENESIS_DELAY=30
export HARDFORK_DELAY=900 #15 minutes enough to take export and gracefully close
export NODE_DELAY=30

export GENESIS_STAKE_TYPE="direct" #'delegated' or 'direct' as in direct stake

#if set to 1, each observer will turn off the antiflooding capability, allowing spam in our network
export OBSERVERS_ANTIFLOOD_DISABLE=0

# If set to 1, this will deploy nodes in a sovereign shard.
# All variables from metashard structure(validators, observers, consensus) should be set to zero and SHARDCOUNT to 1
# For now, make sure that you checkout feat/sovereign branch from mx-chain-deploy repo when using these scripts
export SOVEREIGN_DEPLOY=1

# Shard structure
export SHARDCOUNT=1
export SHARD_VALIDATORCOUNT=2
export SHARD_OBSERVERCOUNT=1
export SHARD_CONSENSUS_SIZE=2

# Metashard structure
export META_VALIDATORCOUNT=0
export META_OBSERVERCOUNT=0
export META_CONSENSUS_SIZE=$META_VALIDATORCOUNT

# ROUND_DURATION_IN_MS is the duration in milliseconds for one round
export ROUND_DURATION_IN_MS=6000

# MULTI_KEY_NODES if set to 1, one observer will be generated on each shard that will handle all generated keys
export MULTI_KEY_NODES=0

# EXTRA_KEYS if set to 1, extra keys will be added to the generated keys
export EXTRA_KEYS=1

# ALWAYS_NEW_CHAINID will generate a fresh new chain ID each time start.sh/config.sh is called
export ALWAYS_NEW_CHAINID=1

# DEFAULT_CHAIN_ID represents the default chain ID
export DEFAULT_CHAIN_ID="local-testnet"

# ROUNDS_PER_EPOCH represents the number of rounds per epoch. If set to 0, it won't override the node's config
export ROUNDS_PER_EPOCH=0

# HYSTERESIS defines the hysteresis value for number of nodes in shard
export HYSTERESIS=0.0

# ALWAYS_NEW_APP_VERSION will set a new version each time the node will be compiled
export ALWAYS_NEW_APP_VERSION=0

# ALWAYS_UPDATE_CONFIGS will re-generate configs (toml + json) each time ./start.sh
# Set this variable to 0 when testing bootstrap from storage or other edge cases where you do not want a fresh new config
# each time.
export ALWAYS_UPDATE_CONFIGS=1

# IP of the seednode
export SEEDNODE_IP="127.0.0.1"

# Ports used by the Nodes
export PORT_SEEDNODE="9999"
export PORT_ORIGIN_OBSERVER="21100"
export PORT_ORIGIN_OBSERVER_REST="10000"
export PORT_ORIGIN_VALIDATOR="21500"
export PORT_ORIGIN_VALIDATOR_REST="9500"

# UI configuration profiles

# Use tmux or not. If set to 1, only 2 terminal windows will be opened, and
# tmux will be used to display the running executables using split windows.
# Recommended. Tmux needs to be installed.
export USETMUX=1

# Log level for the logger in the Node.
export LOGLEVEL="*:DEBUG"


if [ "$TESTNETMODE" == "debug" ]; then
  LOGLEVEL="*:DEBUG,api:INFO"
fi

if [ "$TESTNETMODE" == "trace" ]; then
  LOGLEVEL="*:TRACE"
fi

########################################################################
# Proxy configuration

# Path to mx-chain-proxy-go, branch: master. Default: near mx-chain-go.
export PROXYDIR="$(dirname $MULTIVERSXDIR)/mx-chain-proxy-go/cmd/proxy"
export PROXY=$PROXYDIR/proxy    # Leave unchanged.

export PORT_PROXY="7950"
export PROXY_DELAY=10



########################################################################
# TxGen configuration

# Path to mx-chain-txgen-go. Default: near mx-chain-go.
export TXGENDIR="$(dirname $MULTIVERSXDIR)/mx-chain-txgen-go/cmd/txgen"
export TXGEN=$TXGENDIR/txgen    # Leave unchanged.

export PORT_TXGEN="7951"

export TXGEN_SCENARIOS_LINE='Scenarios = ["basic", "erc20", "esdt"]'

# Number of accounts to be generated by txgen
export NUMACCOUNTS="250"

# Whether txgen should regenerate its accounts when starting, or not.
# Recommended value is 1, but 0 is useful to run the txgen a second time, to
# continue a testing session on the same accounts.
export TXGEN_REGENERATE_ACCOUNTS=1

# COPY_BACK_CONFIGS when set to 1 will copy back the configs and keys to the ./cmd/node/config directory
# in order to have a node in the IDE that can run a node in debug mode but in the same network with the rest of the nodes
# this option greatly helps the debugging process when running a small system test
export COPY_BACK_CONFIGS=0
# SKIP_VALIDATOR_IDX when setting a value greater than -1 will not launch the validator with the provided index
export SKIP_VALIDATOR_IDX=-1
# SKIP_OBSERVER_IDX when setting a value greater than -1 will not launch the observer with the provided index
export SKIP_OBSERVER_IDX=-1

# USE_HARDFORK will prepare the nodes to run the hardfork process, if needed
export USE_HARDFORK=1

# Load local overrides, .gitignored
LOCAL_OVERRIDES="$MULTIVERSXTESTNETSCRIPTSDIR/local.sh"
if [ -f "$LOCAL_OVERRIDES" ]; then
  source "$MULTIVERSXTESTNETSCRIPTSDIR/local.sh"
fi

# Leave unchanged.
let "total_observer_count = $SHARD_OBSERVERCOUNT * $SHARDCOUNT + $META_OBSERVERCOUNT"
export TOTAL_OBSERVERCOUNT=$total_observer_count

# to enable the full archive feature on the observers, please use the --full-archive flag
export EXTRA_OBSERVERS_FLAGS="-operation-mode db-lookup-extension"

# Leave unchanged.
let "total_node_count = $SHARD_VALIDATORCOUNT * $SHARDCOUNT + $META_VALIDATORCOUNT + $TOTAL_OBSERVERCOUNT"
export TOTAL_NODECOUNT=$total_node_count

# VALIDATOR_KEY_PEM_FILE is the pem file name when running single key mode, with all nodes' keys
export VALIDATOR_KEY_PEM_FILE="validatorKey.pem"

# MULTI_KEY_PEM_FILE is the pem file name when running multi key mode, with all managed
export MULTI_KEY_PEM_FILE="allValidatorsKeys.pem"

# EXTRA_KEY_PEM_FILE is the pem file name when running multi key mode, with all extra managed
export EXTRA_KEY_PEM_FILE="extraValidatorsKeys.pem"
