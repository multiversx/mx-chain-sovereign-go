computeFirstSovereignContractAddress() {
    local ADDRESS=$(python3 $SCRIPT_PATH/pyScripts/compute_contract_address.py $WALLET_ADDRESS 0)
    echo $(python3 $TESTNET_DIR/convert_address.py $ADDRESS $ADDRESS_HRP)
}

computeSecondSovereignContractAddress() {
    local ADDRESS=$(python3 $SCRIPT_PATH/pyScripts/compute_contract_address.py $WALLET_ADDRESS 1)
    echo $(python3 $TESTNET_DIR/convert_address.py $ADDRESS $ADDRESS_HRP)
}

getShardOfAddress() {
    echo $(python3 $SCRIPT_PATH/pyScripts/address_shard.py $WALLET_ADDRESS)
}

bech32ToHex() {
    echo $(python3 $SCRIPT_PATH/pyScripts/bech32_to_hex.py $1)
}

hexToBech32() {
    echo $(python3 $SCRIPT_PATH/pyScripts/hex_to_bech32.py $1)
}

updateAndStartBridgeService() {
    python3 $SCRIPT_PATH/pyScripts/bridge_service.py $WALLET $PROXY $ESDT_SAFE_ADDRESS $HEADER_VERIFIER_ADDRESS
}

setGenesisContracts() {
    local ESDT_SAFE_INIT_PARAMS="$(bech32ToHex $FEE_MARKET_ADDRESS_SOVEREIGN)"
    local FEE_MARKET_INIT_PARAMS="$(bech32ToHex $ESDT_SAFE_ADDRESS_SOVEREIGN)@00"
    local ADDRESS=$(python3 $TESTNET_DIR/convert_address.py $WALLET_ADDRESS $ADDRESS_HRP)

    python3 $SCRIPT_PATH/pyScripts/genesis_contracts.py $ADDRESS $SOV_ESDT_SAFE_WASM $ESDT_SAFE_INIT_PARAMS $SOV_FEE_MARKET_WASM $FEE_MARKET_INIT_PARAMS
}

updateSovereignTomlConfigs() {
    python3 $SCRIPT_PATH/pyScripts/update_toml.py $ESDT_SAFE_ADDRESS $ESDT_SAFE_ADDRESS_SOVEREIGN $SOV_CHAIN_PREFIX $USE_ELASTICSEARCH $MAIN_CHAIN_ELASTIC $NATIVE_ESDT $HEADER_VERIFIER_ADDRESS
}

updateNotifierNotarizationRound() {
    python3 $SCRIPT_PATH/pyScripts/notifier_round.py $PROXY $(getShardOfAddress)
}

updateSovereignNodeConfigs() {
    setGenesisContracts

    updateSovereignTomlConfigs

    updateNotifierNotarizationRound
}