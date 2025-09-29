package clob

const (
	ProcessOrderEndpoint = "processOrder"
	CancelOrderEndpoint  = "cancelOrder"
	GetOrderEndpoint     = "getOrder"
	GetDepthEndpoint     = "getDepth"
	MatchOrdersEndpoint  = "matchOrders"
)

// OrderType of the Order
type OrderType string

// Different order types
const (
	TypeMarket    OrderType = "MARKET"
	TypeLimit     OrderType = "LIMIT"
	TypeStopLimit OrderType = "STOP-LIMIT"
)

// Role of the Order
type Role string

// Different order roles
const (
	MAKER Role = "MAKER"
	TAKER Role = "TAKER"
)

// Side of the Order
type Side byte

// Different order sides
const (
	SideBuy  Side = 'B'
	SideSell Side = 'S'
)

// TIF of the Order
type TIF string

// Different order TIF
const (
	GTC TIF = "GTC"
	FOK TIF = "FOK"
	IOC TIF = "IOC"
)