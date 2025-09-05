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
}

startSovereign() {
    updateAndStartBridgeService

    $TESTNET_DIR/sovereignStart.sh
}

stopSovereign() {
    $TESTNET_DIR/stop.sh

    screen -S sovereignBridgeService -X kill
}

stopAndCleanSovereign() {
    stopSovereign

    $TESTNET_DIR/clean.sh
}
