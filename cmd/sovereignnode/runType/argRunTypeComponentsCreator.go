package runType

import (
	"fmt"

	"github.com/multiversx/mx-sdk-abi-go/abi"

	"github.com/multiversx/mx-chain-sovereign-go/cmd/sovereignnode/dataCodec"

	"github.com/multiversx/mx-chain-sovereign-go/config"
	"github.com/multiversx/mx-chain-sovereign-go/factory/runType"
	"github.com/multiversx/mx-chain-sovereign-go/process/block/sovereign/incomingHeader"
)

const (
	separator = "@"
)

// CreateSovereignArgsRunTypeComponents creates the args for run type component
func CreateSovereignArgsRunTypeComponents(
	argsRunType runType.ArgsRunTypeComponents,
	configs config.SovereignConfig,
) (*runType.ArgsSovereignRunTypeComponents, error) {
	runTypeComponentsFactory, err := runType.NewRunTypeComponentsFactory(argsRunType)
	if err != nil {
		return nil, fmt.Errorf("NewRunTypeComponentsFactory failed: %w", err)
	}

	serializer, err := abi.NewSerializer(abi.ArgsNewSerializer{
		PartsSeparator: separator,
	})
	if err != nil {
		return nil, err
	}

	dataCodecHandler, err := dataCodec.NewDataCodec(serializer)
	if err != nil {
		return nil, err
	}

	return &runType.ArgsSovereignRunTypeComponents{
		RunTypeComponentsFactory: runTypeComponentsFactory,
		Config:                   configs,
		DataCodec:                dataCodecHandler,
		TopicsChecker:            incomingHeader.NewTopicsChecker(),
	}, nil
}
