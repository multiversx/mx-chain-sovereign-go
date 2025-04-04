package mempool

import (
	"testing"

	"github.com/multiversx/mx-chain-go/config"
)

func TestMempoolWithSovereignChainSimulator_Selection(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	simulator := startSovereignChainSimulator(t, func(cfg *config.Configs) {})
	defer simulator.Close()

	testSelection(t, simulator)
}

func TestMempoolWithSovereignChainSimulator_Selection_WhenUsersHaveZeroBalance_WithRelayedV3(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	simulator := startSovereignChainSimulator(t, func(cfg *config.Configs) {})
	defer simulator.Close()

	testSelection_WhenUsersHaveZeroBalance_WithRelayedV3(t, simulator)
}

func TestMempoolWithSovereignChainSimulator_Selection_WhenInsufficientBalanceForFee_WithRelayedV3(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	simulator := startSovereignChainSimulator(t, func(cfg *config.Configs) {})
	defer simulator.Close()

	testSelection_WhenInsufficientBalanceForFee_WithRelayedV3(t, simulator)
}

func TestMempoolWithSovereignChainSimulator_Eviction(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	simulator := startSovereignChainSimulator(t, func(cfg *config.Configs) {})
	defer simulator.Close()

	testEviction(t, simulator)
}
