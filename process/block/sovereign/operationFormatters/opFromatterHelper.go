package operationFormatters

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

type opFormatterHelper struct {
	dataCodec     DataCodecHandler
	topicsChecker TopicsCheckerHandler
}

func (op *opFormatterHelper) checkAndGetEventData(event data.EventHandler) (*sovereign.EventData, error) {
	evData, err := op.dataCodec.DeserializeEventData(event.GetData())
	if err != nil {
		return nil, err
	}

	topics := event.GetTopics()
	err = op.topicsChecker.CheckValidity(topics, evData.TransferData)
	if err != nil {
		return nil, err
	}

	return evData, nil
}
