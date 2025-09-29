package clob

import "github.com/gammazero/deque"

// StopBook represents a book of stop orders.
type StopBook struct {
	orders *deque.Deque[*Order]
}

// NewStopBook creates a new instance of the StopBook.
func NewStopBook() *StopBook {
	return &StopBook{
		orders: deque.New[*Order](),
	}
}

// Len returns the number of orders in the book.
func (sb *StopBook) Len() int {
	return sb.orders.Len()
}

// Append appends a new stop order to the book.
func (sb *StopBook) Append(order *Order) {
	sb.orders.PushBack(order)
}

// Remove removes a specific stop order from the book.
func (sb *StopBook) Remove(order *Order) *Order {
	for i := 0; i < sb.orders.Len(); i++ {
		if sb.orders.At(i).GetID() == order.GetID() {
			return sb.orders.Remove(i)
		}
	}
	return nil
}

// Iterate iterates through the stop orders and processes them if the stop price is triggered.
func (sb *StopBook) Iterate(process func(*Order)) {
	for i := 0; i < sb.orders.Len(); i++ {
		order := sb.orders.At(i)
		process(order)
	}
}