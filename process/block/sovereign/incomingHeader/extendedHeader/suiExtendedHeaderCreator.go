package extendedHeader

import (
	"encoding/json"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
	"github.com/multiversx/sui-chain-sovereign-notifier-go/tracker"
)

type suiExtendedHeaderCreator struct {
}

// NewSUIExtendedHeaderCreator will create an SUI chain extended header creator
func NewSUIExtendedHeaderCreator() *suiExtendedHeaderCreator {
	return &suiExtendedHeaderCreator{}
}

// CreateNewExtendedHeader will create an extended header for SUI chain, given the proof.
// For now, proof represents the json marshalled bytes of the header.
// Basic fields like nonce will be assigned to the MVX Header struct.
func (creator *suiExtendedHeaderCreator) CreateNewExtendedHeader(proof []byte) (data.ShardHeaderExtendedHandler, error) {
	suiCheckpoint := &tracker.SUILightCheckpoint{}
	err := json.Unmarshal(proof, suiCheckpoint)
	if err != nil {
		return nil, err
	}

	return &block.ShardHeaderExtended{
		Header: &block.HeaderV2{
			Header: &block.Header{
				Nonce: suiCheckpoint.IncomingNonce,
				Round: suiCheckpoint.IncomingNonce,
				// TODO: MX-17145 maybe extend ShardHeaderExtended from core to return on GetShardID the source chain
				ShardID: uint32(dto.SUI),
			},
		},
		Proof:         proof,
		SourceChainID: dto.SUI,
	}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (creator *suiExtendedHeaderCreator) IsInterfaceNil() bool {
	return creator == nil
}
