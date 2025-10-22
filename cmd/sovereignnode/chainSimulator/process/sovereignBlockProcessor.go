package process

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls/sovereign"
	sovereignBlock "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/sovereign"
	chainSimulatorProcess "github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	"github.com/multiversx/mx-chain-go/process"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
)

type sovereignBlockProcessor struct {
}

// NewSovereignBlockProcessor creates a new block processor for sovereign chain simulator
func NewSovereignBlockProcessor() chainSimulatorProcess.BlocksProcessor {
	return &sovereignBlockProcessor{}
}

// ProcessBlock will create and process a block in sovereign chain simulator
func (sbpf *sovereignBlockProcessor) ProcessBlock(blockProcessor process.BlockProcessor, header data.HeaderHandler) (data.HeaderHandler, data.BodyHandler, error) {
	if check.IfNil(blockProcessor) {
		return nil, nil, process.ErrNilBlockProcessor
	}
	if check.IfNil(header) {
		return nil, nil, process.ErrNilHeaderHandler
	}

	header, block, err := blockProcessor.CreateBlock(header, func() bool {
		return true
	})
	if err != nil {
		return nil, nil, err
	}

	return blockProcessor.ProcessBlock(header, block, func() time.Duration {
		return time.Second
	})
}

// ProcessHeaderProof will process the header proof by updating outgoing op pool with proof signatures
func (sbpf *sovereignBlockProcessor) ProcessHeaderProof(
	header data.HeaderHandler,
	proof data.HeaderProofHandler,
	outGoingOperationsPool sovereignBlock.OutGoingOperationsPool,
) error {
	sovHdr, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in ProcessHeaderProof", process.ErrWrongTypeAssertion)
	}

	for _, outGoingMb := range sovHdr.GetOutGoingMiniBlockHeaderHandlers() {
		mbType := block.OutGoingMBType(outGoingMb.GetChainID()).String()
		extraSigData, found := proof.GetExtraSignatureHandlers()[mbType]
		if !found {
			return fmt.Errorf("%w for type %s in ProcessHeaderProof", bls.ErrExtraSigShareDataNotFound, mbType)
		}

		_, err := sovereign.UpdateBridgeDataWithSignatures(&sovereign.BridgeDataSignatures{
			Hash:      outGoingMb.GetOutGoingOperationsHash(),
			AggSig:    extraSigData.GetAggregatedSignature(),
			LeaderSig: extraSigData.GetLeaderSignature(),
			Bitmap:    proof.GetPubKeysBitmap(),
		}, outGoingOperationsPool)
		if err != nil {
			return err
		}
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sbpf *sovereignBlockProcessor) IsInterfaceNil() bool {
	return sbpf == nil
}
