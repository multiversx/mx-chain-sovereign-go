fund() {
    if [ "$#" -ne 1 ]; then
        echo "Usage: fund <address>"
        return 1
    fi

    local RECEIVER=$(python3 $TESTNET_DIR/convert_address.py $1 $ADDRESS_HRP)
    echo "Funding wallet address $RECEIVER on sovereign chain..."

    local OUTFILE="${OUTFILE_PATH}/get-funds-sovereign.interaction.json"
    mxpy tx new \
        --pem=${WALLET_SOVEREIGN} \
        --proxy=${PROXY_SOVEREIGN} \
        --chain=${CHAIN_ID_SOVEREIGN} \
        --receiver=$RECEIVER \
        --value=100000000000000000000000 \
        --gas-limit=50000 \
        --outfile=${OUTFILE} \
        --wait-result \
        --send

        printTxStatus ${OUTFILE} || return
}

gitPullAllChanges()
{
    pushd .

    # Traverse up to the parent directory of "mx-chain-sovereign-go"
    while [[ ! -d "mx-chain-sovereign-go" && $(pwd) != "/" ]]; do
      cd ..
    done

    # Check if we found the directory
    if [[ ! -d "mx-chain-sovereign-go" ]]; then
      echo "mx-chain-sovereign-go directory not found"
      popd
      return 1
    fi

    echo -e "Pulling changes for mx-chain-sovereign-go..."
    cd mx-chain-sovereign-go
    git pull
    cd ..

    echo -e "Pulling changes for mx-chain-deploy-sovereign-go..."
    cd mx-chain-deploy-sovereign-go
    git pull
    cd ..

    echo -e "Pulling changes for mx-chain-proxy-sovereign-go..."
    cd mx-chain-proxy-sovereign-go
    git pull
    cd ..

    echo -e "Pulling changes for mx-chain-sovereign-bridge-go..."
    pushd .
    cd mx-chain-sovereign-bridge-go
    git pull
    cd cert/cmd/cert
    go build
    ./cert
    popd

    echo -e "Pulling changes for mx-chain-tools-go..."
    pushd .
    cd mx-chain-tools-go
    git pull
    cd elasticreindexer/cmd/indices-creator/
    go build
    popd

    popd

    pip install --upgrade multversx-sdk
    pipx upgrade multiversx-sdk-cli --force
}
