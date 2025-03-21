package extendedHeader

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"

	"github.com/multiversx/mx-chain-go/errors"
)

type emptyMVXHeaderV2Creator struct {
	marshaller marshal.Marshalizer
}

// NewEmptyHeaderV2Creator is able to create empty header v2 instances
func NewEmptyHeaderV2Creator(marshaller marshal.Marshalizer) (*emptyMVXHeaderV2Creator, error) {
	if check.IfNil(marshaller) {
		return nil, data.ErrNilMarshalizer
	}

	return &emptyMVXHeaderV2Creator{
		marshaller: marshaller,
	}, nil
}

// CreateNewExtendedHeader creates a new empty mvx extended header
func (creator *emptyMVXHeaderV2Creator) CreateNewExtendedHeader(proof []byte) (data.ShardHeaderExtendedHandler, error) {
	headerHandler, err := block.GetHeaderFromBytes(creator.marshaller, block.NewEmptyHeaderV2Creator(), proof)
	if err != nil {
		return nil, err
	}

	headerV2, castOk := headerHandler.(*block.HeaderV2)
	if !castOk {
		return nil, fmt.Errorf("%w in emptyMVXHeaderV2Creator.CreateNewExtendedHeader", errors.ErrWrongTypeAssertion)
	}

	return &block.ShardHeaderExtended{
		Header: headerV2,
	}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (creator *emptyMVXHeaderV2Creator) IsInterfaceNil() bool {
	return creator == nil
}
