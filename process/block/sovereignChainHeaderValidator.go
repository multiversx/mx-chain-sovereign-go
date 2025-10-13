package block

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignChainHeaderValidator struct {
	*headerValidator
}

// NewSovereignChainHeaderValidator creates a new sovereign chain header validator
func NewSovereignChainHeaderValidator(
	headerValidator *headerValidator,
) (*sovereignChainHeaderValidator, error) {
	if headerValidator == nil {
		return nil, process.ErrNilHeaderValidator
	}

	schv := &sovereignChainHeaderValidator{
		headerValidator: headerValidator,
	}

	schv.calculateHeaderHashFunc = schv.calculateHeaderHash
	return schv, nil
}

func (schv *sovereignChainHeaderValidator) calculateHeaderHash(headerHandler data.HeaderHandler) ([]byte, error) {
	shardHeaderExtended, isShardHeaderExtended := headerHandler.(*block.ShardHeaderExtended)
	if isShardHeaderExtended {
		if check.IfNil(shardHeaderExtended.Header) {
			return nil, process.ErrNilHeaderHandler
		}

		return core.CalculateHash(schv.marshalizer, schv.hasher, shardHeaderExtended.Header)
	}

	return core.CalculateHash(schv.marshalizer, schv.hasher, headerHandler)
}

// IsHeaderConstructionValid verifies if current header is constructed correctly on top of previous header
func (schv *sovereignChainHeaderValidator) IsHeaderConstructionValid(currHeader, prevHeader data.HeaderHandler) error {
	extendedHdr, isExtendedHeader := currHeader.(data.ShardHeaderExtendedHandler)
	if isExtendedHeader && extendedHdr.GetSourceChainID() == dto.SUI {
		log.Error("SUI")
		return nil
	}

	err := schv.checkHdrRoundAndNonce(currHeader, prevHeader)
	if err != nil {
		return err
	}

	if isETHChainHdr(currHeader) {
		return nil
	}

	return schv.checkHdrHashes(currHeader, prevHeader)
}

func isETHChainHdr(currHeader data.HeaderHandler) bool {
	extendedHdr, isExtendedHeader := currHeader.(data.ShardHeaderExtendedHandler)
	if !isExtendedHeader {
		return false
	}

	return extendedHdr.GetSourceChainID() == dto.ETH
}

// IsInterfaceNil returns if underlying object is true
func (schv *sovereignChainHeaderValidator) IsInterfaceNil() bool {
	return schv == nil
}
