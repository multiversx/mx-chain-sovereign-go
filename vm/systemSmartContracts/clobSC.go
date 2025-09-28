package systemSmartContracts

import (
	"encoding/json"
	"github.com/nikolaydubina/fpdecimal"
)

// clobSC is the smart contract that handles the CLOB logic.
type clobSC struct {
	clob *CLOB
}

// NewClobSC creates a new instance of the clobSC.
func NewClobSC() *clobSC {
	return &clobSC{
		clob: NewCLOB(),
	}
}

// ProcessOrder handles the processing of a new order.
func (sc *clobSC) ProcessOrder(
	orderID string,
	side Side,
	orderType OrderType,
	quantity, price, stop fpdecimal.Decimal,
	tif TIF,
	oco string,
) ([]byte, error) {
	done, err := sc.clob.ProcessOrder(orderID, side, orderType, quantity, price, stop, tif, oco)
	if err != nil {
		return nil, err
	}
	return json.Marshal(done)
}

// CancelOrder handles the cancellation of an existing order.
func (sc *clobSC) CancelOrder(orderID string) ([]byte, error) {
	order := sc.clob.CancelOrder(orderID)
	if order == nil {
		return nil, ErrOrderNotFound
	}
	return json.Marshal(order)
}

// GetOrder retrieves an order by its ID.
func (sc *clobSC) GetOrder(orderID string) ([]byte, error) {
	order := sc.clob.GetOrder(orderID)
	if order == nil {
		return nil, ErrOrderNotFound
	}
	return json.Marshal(order)
}

// GetDepth retrieves the order book depth.
func (sc *clobSC) GetDepth() ([]byte, error) {
	depth := sc.clob.GetDepth()
	return json.Marshal(depth)
}

// LoadState loads the contract's state from storage.
func (sc *clobSC) LoadState(data []byte) error {
	return sc.clob.LoadState(data)
}

// SaveState saves the contract's state to storage.
func (sc *clobSC) SaveState() ([]byte, error) {
	return sc.clob.SaveState()
}