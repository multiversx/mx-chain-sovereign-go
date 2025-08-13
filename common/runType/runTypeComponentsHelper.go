package runType

import (
	"github.com/multiversx/mx-chain-core-go/core"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/data"
)

// ReadInitialAccounts returns the genesis accounts from a file
func ReadInitialAccounts(filePath string) ([]genesis.InitialAccountHandler, error) {
	initialAccounts := make([]*data.InitialAccount, 0)
	err := core.LoadJsonFile(&initialAccounts, filePath)
	if err != nil {
		return nil, err
	}

	var accounts []genesis.InitialAccountHandler
	for _, ia := range initialAccounts {
		accounts = append(accounts, ia)
	}

	return accounts, nil
}

// CreateSovereignProposedInitialHeader will create the initial sovereign basic header handler data that is going
// to be proposed, signed and broadcast by the leader in andromeda sovereign consensus. Leader's signature in block
// will be applied to this header
func CreateSovereignProposedInitialHeader(header coreData.HeaderHandler) coreData.HeaderHandler {
	return &block.SovereignChainHeader{
		Header: &block.Header{
			Nonce:        header.GetNonce(),
			PrevHash:     header.GetPrevHash(),
			PrevRandSeed: header.GetPrevRandSeed(),
			RandSeed:     header.GetRandSeed(),
			ShardID:      header.GetShardID(),
			TimeStamp:    header.GetTimeStamp(),
			Round:        header.GetRound(),
			Epoch:        header.GetEpoch(),
			ChainID:      header.GetChainID(),
		},
		IsStartOfEpoch: header.IsStartOfEpochBlock(),
	}
}
