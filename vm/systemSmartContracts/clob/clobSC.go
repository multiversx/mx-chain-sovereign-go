package clob

import (
	"math/big"
)

// clobSC is the smart contract that handles the CLOB logic.
type clobSC struct {
	clob *CLOB
}

// NewClobSC creates a new instance of the clobSC.
func NewClobSC(clob *CLOB) *clobSC {
	return &clobSC{
		clob: clob,
	}
}

// MatchOrders matches orders in the order book.
func (sc *clobSC) MatchOrders() (*Done, error) {
	dones, err := sc.clob.MatchOrders()
	if err != nil {
		return nil, err
	}
	if len(dones) > 0 {
		return dones[0], nil
	}
	return nil, nil
}

// ProcessOrder handles the processing of a new order.
func (sc *clobSC) ProcessOrder(
	orderID string,
	side Side,
	orderType OrderType,
	quantity, price, stop *big.Float,
	tif TIF,
	oco string,
) (*Done, error) {
	return sc.clob.ProcessOrder(orderID, side, orderType, quantity, price, stop, tif, oco)
}

// CancelOrder handles the cancellation of an existing order.
func (sc *clobSC) CancelOrder(orderID string) (*Order, error) {
	return sc.clob.CancelOrder(orderID)
}

// GetOrder retrieves an order by its ID.
func (sc *clobSC) GetOrder(orderID string) (*Order, error) {
	return sc.clob.GetOrder(orderID)
}

// GetDepth retrieves the order book depth.
func (sc *clobSC) GetDepth() *Depth {
	return sc.clob.GetDepth()
}

// LoadState loads the contract's state from storage.
func (sc *clobSC) LoadState(data []byte) error {
	return sc.clob.LoadState(data)
}

// SaveState saves the contract's state to storage.
func (sc *clobSC) SaveState() ([]byte, error) {
	return sc.clob.SaveState()
}