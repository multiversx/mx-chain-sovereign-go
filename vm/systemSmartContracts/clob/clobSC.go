package clob

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

// ProcessOrder handles the processing of a new order.
func (sc *clobSC) ProcessOrder(
	orderID string,
	side Side,
	orderType OrderType,
	quantity, price, stop Decimal,
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
func (sc *clobSC) GetDepth() (*Depth, error) {
	return sc.clob.GetDepth(), nil
}

// LoadState loads the contract's state from storage.
func (sc *clobSC) LoadState(data []byte) error {
	return sc.clob.LoadState(data)
}

// SaveState saves the contract's state to storage.
func (sc *clobSC) SaveState() ([]byte, error) {
	return sc.clob.SaveState()
}