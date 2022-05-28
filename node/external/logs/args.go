package logs

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

// ArgsNewFeeComputer holds the arguments for constructing a logsFacade
type ArgsNewLogsFacade struct {
	StorageService  dataRetriever.StorageService
	Marshalizer     marshal.Marshalizer
	PubKeyConverter core.PubkeyConverter
}

func (args *ArgsNewLogsFacade) check() error {
	if check.IfNil(args.StorageService) {
		return core.ErrNilStore
	}
	if check.IfNil(args.Marshalizer) {
		return core.ErrNilMarshalizer
	}
	if check.IfNil(args.PubKeyConverter) {
		return ErrNilPubkeyConverter
	}

	return nil
}
