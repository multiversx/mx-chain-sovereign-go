package extendedHeader

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
	"github.com/multiversx/mx-chain-core-go/marshal"

	"github.com/multiversx/mx-chain-go/errors"
)

type emptyMVXShardExtendedCreator struct {
	marshaller           marshal.Marshalizer
	headerV2BlockCreator block.EmptyBlockCreator
}

// NewEmptyMVXShardExtendedCreator is able to create empty mvx header v2 instances from proofs
func NewEmptyMVXShardExtendedCreator(marshaller marshal.Marshalizer) (*emptyMVXShardExtendedCreator, error) {
	if check.IfNil(marshaller) {
		return nil, data.ErrNilMarshalizer
	}

	return &emptyMVXShardExtendedCreator{
		marshaller:           marshaller,
		headerV2BlockCreator: block.NewEmptyHeaderV2Creator(),
	}, nil
}

// CreateNewExtendedHeader creates a new empty extended header from a MultiversX chain proof
func (creator *emptyMVXShardExtendedCreator) CreateNewExtendedHeader(proof []byte) (data.ShardHeaderExtendedHandler, error) {
	headerHandler, err := block.GetHeaderFromBytes(creator.marshaller, creator.headerV2BlockCreator, proof)
	if err != nil {
		return nil, err
	}

	headerV2, castOk := headerHandler.(*block.HeaderV2)
	if !castOk {
		return nil, fmt.Errorf("%w in emptyMVXShardExtendedCreator.CreateNewExtendedHeader", errors.ErrWrongTypeAssertion)
	}

	return &block.ShardHeaderExtended{
		Header:        headerV2,
		Proof:         proof,
		SourceChainID: dto.MVX,
	}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (creator *emptyMVXShardExtendedCreator) IsInterfaceNil() bool {
	return creator == nil
}
