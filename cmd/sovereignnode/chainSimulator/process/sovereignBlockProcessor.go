package process

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	sovereignBlock "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/sovereign"
	"github.com/multiversx/mx-chain-go/errors"
	chainSimulatorProcess "github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	"github.com/multiversx/mx-chain-go/process"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
)

type bridgeDataSignatures struct {
	hash      []byte
	aggSig    []byte
	leaderSig []byte
	bitmap    []byte
}

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
		mbType := block.OutGoingMBType(outGoingMb.GetOutGoingMBTypeInt32()).String()
		extraSigData, found := proof.GetExtraSignatureHandlers()[mbType]
		if !found {
			return fmt.Errorf("%w for type %s in ProcessHeaderProof", bls.ErrExtraSigShareDataNotFound, mbType)
		}

		err := sbpf.updateBridgeDataWithSignatures(&bridgeDataSignatures{
			hash:      outGoingMb.GetOutGoingOperationsHash(),
			aggSig:    extraSigData.GetAggregatedSignature(),
			leaderSig: extraSigData.GetLeaderSignature(),
			bitmap:    proof.GetPubKeysBitmap(),
		}, outGoingOperationsPool)
		if err != nil {
			return err
		}
	}

	return nil
}

func (sbpf *sovereignBlockProcessor) updateBridgeDataWithSignatures(
	bridgeDataSigs *bridgeDataSignatures,
	outGoingOperationsPool sovereignBlock.OutGoingOperationsPool,
) error {
	hash := bridgeDataSigs.hash
	currBridgeData := outGoingOperationsPool.Get(hash)
	if currBridgeData == nil {
		return fmt.Errorf("%w for hash: %s",
			errors.ErrOutGoingOperationsNotFound, hex.EncodeToString(hash))
	}

	currBridgeData.LeaderSignature = bridgeDataSigs.leaderSig
	currBridgeData.AggregatedSignature = bridgeDataSigs.aggSig
	currBridgeData.PubKeysBitmap = bridgeDataSigs.bitmap

	outGoingOperationsPool.Delete(hash)
	outGoingOperationsPool.Add(currBridgeData)
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sbpf *sovereignBlockProcessor) IsInterfaceNil() bool {
	return sbpf == nil
}
