IMAGE_NAME="multiversx/sov-observer"
CONTAINER_NAME="sov-observer"

prepareObserver() {
    manualUpdateConfigFile #update config file

    local DOCKER_IMAGE=""

    if [ -n "$1" ]; then
        DOCKER_IMAGE=$1
    else
        case $CHAIN_ID in
            "1")
                TAG=$(curl -s https://registry.hub.docker.com/v2/repositories/multiversx/chain-mainnet/tags | jq -r '.results[0].name')
                DOCKER_IMAGE="multiversx/chain-mainnet:$TAG"
                ;;
            "D")
                TAG=$(curl -s https://registry.hub.docker.com/v2/repositories/multiversx/chain-devnet/tags | jq -r '.results[0].name')
                DOCKER_IMAGE="multiversx/chain-devnet:$TAG"
                ;;
            "T")
                TAG=$(curl -s https://registry.hub.docker.com/v2/repositories/multiversx/chain-testnet/tags | jq -r '.results[0].name')
                DOCKER_IMAGE="multiversx/chain-testnet:$TAG"
                ;;
        esac
    fi

    local LINE="FROM $DOCKER_IMAGE"
    sed -i "1s,.*,${LINE}," "$SCRIPT_PATH/observer/shard-observer" # replace first line with the docker image

    # build docker image
    docker image build . -t $IMAGE_NAME -f $SCRIPT_PATH/observer/shard-observer
    # create docker container
    local SHARD=$(getShardOfAddress)
    docker create -p 8083:8080 -p 22111:22111 --name $CONTAINER_NAME $IMAGE_NAME --destination-shard-as-observer=$SHARD
}

deployObserver() {
    docker start $CONTAINER_NAME
}

stopObserver() {
    docker stop $CONTAINER_NAME
}

cleanObserver() {
    docker remove $CONTAINER_NAME
}
