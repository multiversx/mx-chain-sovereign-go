SOV_CHAIN_PREFIX=j8x
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
    ARGS_FILE=$(eval echo "${PHASE_ONE_ARGS_FILE}")
    sed -e "s/\$SOV_CHAIN_PREFIX/$HEX_CHAIN_ID/g" \
        -e "s/\$SHARD_VALIDATORCOUNT/$SHARD_VALIDATORCOUNT/g" config/deployArguments/phaseOneArgs > ${ARGS_FILE}

    local OUTFILE="${OUTFILE_PATH}/deploy-phase-one.interaction.json"
    mxpy contract call ${SOVEREIGN_FORGE_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --gas-limit=30000000 \
        --abi=$(eval echo ${SOVEREIGN_FORGE_ABI}) \
        --function="deployPhaseOne" \
        --arguments-file ${ARGS_FILE}\
        --outfile=${OUTFILE} \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE} || return

    CHAIN_CONFIG_ADDRESS=$(readSovereignContract $CHAIN_CONFIG_INDEX)
}

deployPhaseTwo() {
    echo "Deploying phase two..."

    local OUTFILE="${OUTFILE_PATH}/deploy-phase-two.interaction.json"
    mxpy contract call ${SOVEREIGN_FORGE_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --gas-limit=30000000 \
        --function="deployPhaseTwo" \
        --outfile=${OUTFILE} \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE} || return

    ESDT_SAFE_ADDRESS=$(readSovereignContract $ESDT_SAFE_INDEX)
    ESDT_SAFE_ADDRESS_SOVEREIGN=$(computeFirstSovereignContractAddress)

    echo "Registering native ESDT token..."

    local HEX_TICKER=$(echo -n "$NATIVE_ESDT_TICKER" | xxd -p)
    local HEX_NAME=$(echo -n "$NATIVE_ESDT_NAME" | xxd -p)
    ARGS_FILE=$(eval echo "${REGISTER_NATIVE_ESDT_ARGS_FILE}")
    sed -e "s/\$NATIVE_ESDT_TICKER/$HEX_TICKER/g" \
        -e "s/\$NATIVE_ESDT_NAME/$HEX_NAME/g" config/deployArguments/registerNativeToken > ${ARGS_FILE}

    local OUTFILE="${OUTFILE_PATH}/register-native-token.interaction.json"
    mxpy contract call ${SOVEREIGN_FORGE_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --gas-limit=100000000 \
        --abi=$(eval echo ${SOVEREIGN_FORGE_ABI}) \
        --function="registerNativeToken" \
        --arguments-file ${ARGS_FILE}\
        --value=${ESDT_ISSUE_COST} \
        --outfile=${OUTFILE} \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE} || return

    local NATIVE_ESDT_HEX=$(readNativeToken)
    NATIVE_ESDT=$(hex_to_string "$NATIVE_ESDT_HEX")
    echo "Native ESDT Token: ${NATIVE_ESDT}"
}

deployPhaseThree() {
    echo "Deploying phase three..."

    local OUTFILE="${OUTFILE_PATH}/deploy-phase-three.interaction.json"
    mxpy contract call ${SOVEREIGN_FORGE_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --gas-limit=30000000 \
        --function="deployPhaseThree" \
        --arguments \
            0x00 \
        --outfile=${OUTFILE} \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE} || return

    FEE_MARKET_ADDRESS=$(readSovereignContract $FEE_MARKET_INDEX)
    FEE_MARKET_ADDRESS_SOVEREIGN=$(computeSecondSovereignContractAddress)
}

deployPhaseFour() {
    echo "Deploying phase four..."

    local OUTFILE="${OUTFILE_PATH}/deploy-phase-four.interaction.json"
    mxpy contract call ${SOVEREIGN_FORGE_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --gas-limit=30000000 \
        --function="deployPhaseFour" \
        --outfile=${OUTFILE} \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE} || return

    HEADER_VERIFIER_ADDRESS=$(readSovereignContract $HEADER_VERIFIER_INDEX)
}

registerBLSKeys() {
    echo "Register validator BLS keys in main chain..."
    checkVariables CHAIN_CONFIG_ADDRESS || return

    BLS_PUB_KEYS=$(python3 $SCRIPT_PATH/pyScripts/read_bls_keys.py)

    for BLS_KEY in $BLS_PUB_KEYS; do
        local OUTFILE="${OUTFILE_PATH}/register-${BLS_KEY}.interaction.json"
        mxpy contract call ${CHAIN_CONFIG_ADDRESS} \
            --pem=${WALLET} \
            --proxy=${PROXY} \
            --gas-limit=90000000 \
            --function="register" \
            --arguments \
                ${BLS_KEY} \
            --outfile=${OUTFILE} \
            --wait-result \
            --send || return

        printTxStatus ${OUTFILE} || return
    done
}

completeSetupPhase() {
    echo "Completing setup phase..."

    local OUTFILE="${OUTFILE_PATH}/complete-setup-phase.interaction.json"
    mxpy contract call ${SOVEREIGN_FORGE_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --gas-limit=90000000 \
        --function="completeSetupPhase" \
        --outfile=${OUTFILE} \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE} || return
}

readNativeToken() {
    checkVariables ESDT_SAFE_ADDRESS || return

    mxpy contract query ${ESDT_SAFE_ADDRESS} \
        --proxy=${PROXY} \
        --function="getNativeToken"
}

HEADER_VERIFIER_INDEX=2
ESDT_SAFE_INDEX=3
FEE_MARKET_INDEX=4
CHAIN_CONFIG_INDEX=6
readSovereignContract() {
    if [ "$#" -ne 1 ]; then
        echo "Usage: ${FUNCNAME[0]} <contract_index>"
        return 1
    fi

    local DATA=$(mxpy contract query ${SOVEREIGN_FORGE_ADDRESS} \
        --proxy=${PROXY} \
        --abi=$(eval echo ${SOVEREIGN_FORGE_ABI}) \
        --function="getDeployedSovereignContracts" \
        --arguments=str:${SOV_CHAIN_PREFIX})

    local HEX_ADDRESS=$(echo "$DATA" | jq -r \
        --argjson id "$1" '
        .[][]
        | select(.id.__discriminant__ == $id)
        | .address
        ')

    echo $(hexToBech32 $HEX_ADDRESS)
}
