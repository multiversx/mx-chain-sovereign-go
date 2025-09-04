generateChainId() {
    if [ -z "$1" ]; then
        echo $(generateRandomChainId)
    else
        echo $1
    fi
}