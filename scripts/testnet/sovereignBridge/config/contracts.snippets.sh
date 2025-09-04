downloadCrossChainContracts() {
    echo "Downloading cross-chain contracts..."

    mkdir -p $(eval echo "${CONTRACTS_DIRECTORY}")
    version=$(basename `curl -s https://github.com/multiversx/mx-sovereign-sc/releases/latest -I | grep location | awk -F"https:/" '{print $2}' | tr -d "\r"`)
    wget -O $(eval echo ${SOVEREIGN_FORGE_ABI}) https://github.com/multiversx/mx-sovereign-sc/releases/download/${version}/sovereign-forge.abi.json
    wget -O $(eval echo ${SOV_ESDT_SAFE_WASM}) https://github.com/multiversx/mx-sovereign-sc/releases/download/${version}/sov-esdt-safe.wasm
    wget -O $(eval echo ${SOV_FEE_MARKET_WASM}) https://github.com/multiversx/mx-sovereign-sc/releases/download/${version}/sov-fee-market.wasm
}
