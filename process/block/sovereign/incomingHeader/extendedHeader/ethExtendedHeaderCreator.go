package extendedHeader

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
)

type ethExtendedHeaderCreator struct {
}

// NewETHExtendedHeaderCreator will create an ETH chain extended header creator
func NewETHExtendedHeaderCreator() *ethExtendedHeaderCreator {
	return &ethExtendedHeaderCreator{}
}

// CreateNewExtendedHeader will create an extended header for ETH chain, given the proof.
// For now, proof represents the json marshalled bytes of the header.
// Basic fields like nonce will be assigned to the MVX Header struct.
func (creator *ethExtendedHeaderCreator) CreateNewExtendedHeader(proof []byte) (data.ShardHeaderExtendedHandler, error) {
	ethHeader := &types.Header{}
	err := json.Unmarshal(proof, ethHeader)
	if err != nil {
		return nil, err
	}

	return &block.ShardHeaderExtended{
		Header: &block.HeaderV2{
			Header: &block.Header{
				Nonce: ethHeader.Number.Uint64(),
				Round: ethHeader.Number.Uint64(),
				// TODO: Here, Think here maybe extend from core ShardHeaderExtended to return GetShardID as source chain for mvx as well?
				ShardID: uint32(dto.ETH),
			},
		},
		Proof:         proof,
		SourceChainID: dto.ETH,
	}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (creator *ethExtendedHeaderCreator) IsInterfaceNil() bool {
	return creator == nil
}
