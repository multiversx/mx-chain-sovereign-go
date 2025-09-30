SOV_CHAIN_PREFIX=
NATIVE_ESDT=
CHAIN_CONFIG_ADDRESS=
ESDT_SAFE_ADDRESS=
ESDT_SAFE_ADDRESS_SOVEREIGN=
FEE_MARKET_ADDRESS=
FEE_MARKET_ADDRESS_SOVEREIGN=
HEADER_VERIFIER_ADDRESS=

deployPhaseOne() {
    echo "Deploying phase one..."

    local HEX_CHAIN_ID=$(echo -n "$SOV_CHAIN_PREFIX" | xxd -p)
    sed -e "s/\$SOV_CHAIN_PREFIX/$HEX_CHAIN_ID/g" \
        -e "s/\$SHARD_VALIDATORCOUNT/$SHARD_VALIDATORCOUNT/g" config/deployArguments/phaseOneArgs > "$PHASE_ONE_ARGS_FILE"

    local OUTFILE="$OUTFILE_PATH/deploy-phase-one.interaction.json"
    mxpy contract call $SOVEREIGN_FORGE_ADDRESS \
        --pem "$WALLET" \
        --proxy "$PROXY" \
        --gas-limit 25000000 \
        --abi "$SOVEREIGN_FORGE_ABI" \
        --function "deployPhaseOne" \
        --arguments-file "$PHASE_ONE_ARGS_FILE" \
        --outfile "$OUTFILE" \
        --wait-result \
        --send || return

    printTxStatus "$OUTFILE" || return

    CHAIN_CONFIG_ADDRESS=$(readContractAddress "$CHAIN_CONFIG_INDEX")
    echo "chain-config contract: ${CHAIN_CONFIG_ADDRESS}"
}

deployPhaseTwo() {
    echo "Deploying phase two..."

    local OUTFILE="$OUTFILE_PATH/deploy-phase-two.interaction.json"
    mxpy contract call $SOVEREIGN_FORGE_ADDRESS \
        --pem "$WALLET" \
        --proxy "$PROXY" \
        --gas-limit 30000000 \
        --function "deployPhaseTwo" \
        --outfile "$OUTFILE" \
        --wait-result \
        --send || return

    printTxStatus "$OUTFILE" || return

    ESDT_SAFE_ADDRESS=$(readContractAddress $ESDT_SAFE_INDEX)
    echo "mvx-esdt-safe contract: $ESDT_SAFE_ADDRESS"
    ESDT_SAFE_ADDRESS_SOVEREIGN=$(computeFirstSovereignContractAddress)
    echo "sov-esdt-safe contract: $ESDT_SAFE_ADDRESS_SOVEREIGN"

    echo "Registering native ESDT token..."

    local OUTFILE="$OUTFILE_PATH/register-native-token.interaction.json"
    mxpy contract call $ESDT_SAFE_ADDRESS \
        --pem "$WALLET" \
        --proxy "$PROXY" \
        --gas-limit 80000000 \
        --function "registerNativeToken" \
        --arguments \
            str:"$NATIVE_ESDT_TICKER" \
            str:"$NATIVE_ESDT_NAME" \
        --value $ESDT_ISSUE_COST \
        --outfile "$OUTFILE" \
        --wait-result \
        --send || return

    printTxStatus "$OUTFILE" || return

    NATIVE_ESDT=$(readNativeESDT)
    echo "Native ESDT Token: $NATIVE_ESDT"
}

deployPhaseThree() {
    echo "Deploying phase three..."

    local OUTFILE="$OUTFILE_PATH/deploy-phase-three.interaction.json"
    mxpy contract call $SOVEREIGN_FORGE_ADDRESS \
        --pem "$WALLET" \
        --proxy "$PROXY" \
        --gas-limit 30000000 \
        --function "deployPhaseThree" \
        --arguments \
            0x00 \
        --outfile "$OUTFILE" \
        --wait-result \
        --send || return

    printTxStatus "$OUTFILE" || return

    FEE_MARKET_ADDRESS=$(readContractAddress $FEE_MARKET_INDEX)
    echo "mvx-fee-market contract: $FEE_MARKET_ADDRESS"
    FEE_MARKET_ADDRESS_SOVEREIGN=$(computeSecondSovereignContractAddress)
    echo "sov-fee-market contract: $FEE_MARKET_ADDRESS_SOVEREIGN"
}

deployPhaseFour() {
    echo "Deploying phase four..."

    local OUTFILE="$OUTFILE_PATH/deploy-phase-four.interaction.json"
    mxpy contract call $SOVEREIGN_FORGE_ADDRESS \
        --pem "$WALLET" \
        --proxy "$PROXY" \
        --gas-limit 25000000 \
        --function "deployPhaseFour" \
        --outfile "$OUTFILE" \
        --wait-result \
        --send || return

    printTxStatus "$OUTFILE" || return

    HEADER_VERIFIER_ADDRESS=$(readContractAddress $HEADER_VERIFIER_INDEX)
    echo "header-verifier contract: $HEADER_VERIFIER_ADDRESS"
}

registerBLSKeys() {
    echo "Registering validator BLS keys in main chain..."
    checkVariables CHAIN_CONFIG_ADDRESS || return

    BLS_PUB_KEYS=$(python3 "$SCRIPT_PATH/pyScripts/read_bls_keys.py")

    for BLS_KEY in ${BLS_PUB_KEYS}; do
        local OUTFILE="$OUTFILE_PATH/register-${BLS_KEY}.interaction.json"
        mxpy contract call $CHAIN_CONFIG_ADDRESS \
            --pem "$WALLET" \
            --proxy "$PROXY" \
            --gas-limit 20000000 \
            --function "register" \
            --arguments \
                "$BLS_KEY" \
            --outfile "$OUTFILE" \
            --wait-result \
            --send || return

        printTxStatus "$OUTFILE" || return
    done
}

completeSetupPhase() {
    echo "Completing setup phase..."

    local OUTFILE="$OUTFILE_PATH/complete-setup-phase.interaction.json"
    mxpy contract call $SOVEREIGN_FORGE_ADDRESS \
        --pem "$WALLET" \
        --proxy "$PROXY" \
        --gas-limit 60000000 \
        --function "completeSetupPhase" \
        --outfile "$OUTFILE" \
        --wait-result \
        --send || return

    printTxStatus "$OUTFILE" || return
}

readNativeESDT() {
    checkVariables ESDT_SAFE_ADDRESS || return

    local NATIVE_ESDT_HEX=$(mxpy contract query $ESDT_SAFE_ADDRESS \
        --proxy "$PROXY" \
        --function "getNativeToken")

    echo $(hexToString "$NATIVE_ESDT_HEX")
}

HEADER_VERIFIER_INDEX=2
ESDT_SAFE_INDEX=3
FEE_MARKET_INDEX=4
CHAIN_CONFIG_INDEX=6
readContractAddress() {
    if [ "$#" -ne 1 ]; then
        echo "Usage: ${FUNCNAME[0]} <contract_index>"
        return 1
    fi

    local DATA=$(mxpy contract query $SOVEREIGN_FORGE_ADDRESS \
        --proxy "$PROXY" \
        --abi "$SOVEREIGN_FORGE_ABI" \
        --function "getDeployedSovereignContracts" \
        --arguments str:"$SOV_CHAIN_PREFIX")

    local HEX_ADDRESS=$(echo "$DATA" | jq -r \
        --argjson id "$1" '
        .[][]
        | select(.id.__discriminant__ == $id)
        | .address
        ')

    echo $(hexToBech32 $HEX_ADDRESS)
}

unpauseEsdtSafeContractSovereign() {
    echo "Unpausing sovereign deposits..."

    local OUTFILE="$OUTFILE_PATH/unpause.interaction.json"
    mxpy contract call $ESDT_SAFE_ADDRESS_SOVEREIGN \
        --pem "$WALLET" \
        --proxy "$PROXY_SOVEREIGN" \
        --gas-limit 10000000 \
        --function "unpause" \
        --outfile "$OUTFILE" \
        --wait-result \
        --send || return

    printTxStatus "$OUTFILE" "$PROXY_SOVEREIGN" || return
}
