createAndDeployMainChainObserver() {
    prepareObserver

    createObserver

    deployObserver
}

stopAndCleanMainChainObserver() {
    stopObserver

    cleanObserver
}

deploySovereignWithCrossChainContracts() {
    SOV_CHAIN_PREFIX=$(generateChainId $1)
    echo "Sovereign chain ID: $SOV_CHAIN_PREFIX"

    deployPhaseOne || return

    deployPhaseTwo || return

    deployPhaseThree || return

    deployPhaseFour || return

    updateSovereignNodeConfigs

    $TESTNET_DIR/config.sh

    registerBLSKeys || return

    completeSetupPhase

    startSovereign

    fund $WALLET_ADDRESS

    unpauseEsdtSafeContractSovereign
}

startSovereign() {
    updateAndStartBridgeService

    local START_TIME=$(generateStartTime)
    updateJSONValue "$TESTNETDIR/node/config/nodesSetup.json" "startTime" $START_TIME
    $TESTNET_DIR/sovereignStart.sh

    waitUntilStartTime $START_TIME
}

stopSovereign() {
    $TESTNET_DIR/stop.sh

    screen -S sovereignBridgeService -X kill
}

stopAndCleanSovereign() {
    stopSovereign

    $TESTNET_DIR/clean.sh
}
