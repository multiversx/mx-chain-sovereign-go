package enablers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go/config"
)

const (
	conversionBase    = 10
	bitConversionSize = 64

	disableAsyncCallV1 = "DisableAsyncCallV1"
)

type enableRoundsHandler struct {
	*roundFlagsHolder
}

// NewEnableRoundsHandler creates a new enable rounds handler instance
func NewEnableRoundsHandler(args config.RoundConfig) (*enableRoundsHandler, error) {
	disableAsyncCallV1, err := getRoundConfig(args, disableAsyncCallV1)
	if err != nil {
		return nil, err
	}

	return &enableRoundsHandler{
		roundFlagsHolder: &roundFlagsHolder{
			disableAsyncCallV1: disableAsyncCallV1,
		},
	}, nil
}

func getRoundConfig(args config.RoundConfig, configName string) (*roundFlag, error) {
	activationRound, found := args.RoundActivations[configName]
	if !found {
		return nil, fmt.Errorf("%w for config %s", errMissingRoundActivation, configName)
	}

	round, err := strconv.ParseUint(activationRound.Round, conversionBase, bitConversionSize)
	if err != nil {
		return nil, fmt.Errorf("%w while trying to convert %s for the round config %s",
			err, activationRound.Round, configName)
	}

	log.Debug("loaded round config",
		"name", configName, "round", round,
		"options", fmt.Sprintf("[%s]", strings.Join(activationRound.Options, ", ")))

	return &roundFlag{
		Flag:    &atomic.Flag{},
		round:   round,
		options: activationRound.Options,
	}, nil
}

// CheckRound should be called whenever a new round is known. It will trigger the updating of all containing round flags
func (handler *enableRoundsHandler) CheckRound(round uint64) {
	handler.disableAsyncCallV1.SetValue(handler.disableAsyncCallV1.round <= round)
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *enableRoundsHandler) IsInterfaceNil() bool {
	return handler == nil
}
