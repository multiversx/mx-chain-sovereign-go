package sharding

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"

	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
)

// SovereignNodesSetup holds the sovereign nodes setup
type SovereignNodesSetup struct {
	*NodesSetup
}

// SovereignNodesSetupArgs is a struct placeholder for sovereign nodes setup args
type SovereignNodesSetupArgs struct {
	NodesConfig              config.NodesConfig
	AddressPubKeyConverter   core.PubkeyConverter
	ValidatorPubKeyConverter core.PubkeyConverter
	ChainParametersProvider  ChainParametersHandler
}

// NewSovereignNodesSetup  creates a new decoded sovereign nodes structure from json config file
func NewSovereignNodesSetup(args *SovereignNodesSetupArgs) (*SovereignNodesSetup, error) {
	if check.IfNil(args.AddressPubKeyConverter) {
		return nil, fmt.Errorf("%w for addressPubkeyConverter in NewSovereignNodesSetup", ErrNilPubkeyConverter)
	}
	if check.IfNil(args.ValidatorPubKeyConverter) {
		return nil, fmt.Errorf("%w for validatorPubkeyConverter in NewSovereignNodesSetup", ErrNilPubkeyConverter)
	}
	if check.IfNil(args.ChainParametersProvider) {
		return nil, fmt.Errorf("%w in NewSovereignNodesSetup", ErrNilChainParametersProvider)
	}

	genesisParams, err := args.ChainParametersProvider.ChainParametersForEpoch(0)
	if err != nil {
		return nil, fmt.Errorf("NewSovereignNodesSetup: %w while fetching parameters for epoch 0", err)
	}

	nodes := &NodesSetup{
		addressPubkeyConverter:   args.AddressPubKeyConverter,
		validatorPubkeyConverter: args.ValidatorPubKeyConverter,
		genesisChainParameters:   genesisParams,
	}

	initNodesSetup(nodes, args.NodesConfig)

	sovereignNodes := &SovereignNodesSetup{
		NodesSetup: nodes,
	}

	err = sovereignNodes.processSovereignConfig()
	if err != nil {
		return nil, err
	}

	sovereignNodes.processSovereignShardAssignment()
	sovereignNodes.createSovereignInitialNodesInfo()

	sovereignNodes.nrOfMetaChainNodes = 0
	return sovereignNodes, nil
}

func (ns *SovereignNodesSetup) processSovereignConfig() error {
	var err error

	ns.nrOfNodes = 0
	ns.nrOfMetaChainNodes = 0
	ns.numberOfShards = 1
	err = ns.processInitialNodes()
	if err != nil {
		return err
	}

	if ns.genesisChainParameters.ShardConsensusGroupSize < 1 {
		return ErrNegativeOrZeroConsensusGroupSize
	}
	if ns.genesisChainParameters.ShardMinNumNodes < ns.genesisChainParameters.ShardConsensusGroupSize {
		return ErrMinNodesPerShardSmallerThanConsensusSize
	}
	if ns.nrOfNodes < ns.genesisChainParameters.ShardMinNumNodes {
		return ErrNodesSizeSmallerThanMinNoOfNodes
	}

	if ns.genesisChainParameters.MetachainMinNumNodes != 0 || ns.genesisChainParameters.MetachainConsensusGroupSize != 0 {
		return fmt.Errorf("%w, min nodes and consensus size should be set to zero", errSovereignInvalidMetaConsensusSize)
	}

	return nil
}

func (ns *SovereignNodesSetup) processSovereignShardAssignment() {
	for id := uint32(0); id < ns.nrOfNodes; id++ {
		// consider only nodes with valid public key
		if ns.InitialNodes[id].pubKey != nil {
			ns.InitialNodes[id].assignedShard = core.SovereignChainShardId
			ns.InitialNodes[id].eligible = true
		}
	}
}

func (ns *SovereignNodesSetup) createSovereignInitialNodesInfo() {
	ns.eligible = make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, ns.numberOfShards)
	ns.waiting = make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, 0)
	for _, in := range ns.InitialNodes {
		if in.pubKey != nil && in.address != nil {
			ni := &nodeInfo{
				assignedShard: in.assignedShard,
				eligible:      in.eligible,
				pubKey:        in.pubKey,
				address:       in.address,
				initialRating: in.initialRating,
			}
			ns.eligible[in.assignedShard] = append(ns.eligible[in.assignedShard], ni)
		}
	}
}
