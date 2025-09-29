package clob

import "github.com/gammazero/deque"

// OrderQueue represents a queue of orders for a specific price level.
type OrderQueue struct {
	price  string
	orders *deque.Deque[*Order]
}

// NewOrderQueue creates a new instance of the OrderQueue.
func NewOrderQueue(price string) *OrderQueue {
	return &OrderQueue{
		price:  price,
		orders: deque.New[*Order](),
	}
}

// Len returns the number of orders in the queue.
func (oq *OrderQueue) Len() int {
	return oq.orders.Len()
}

// Front returns the first order in the queue without removing it.
func (oq *OrderQueue) Front() *Order {
	if oq.orders.Len() == 0 {
		return nil
	}
	return oq.orders.Front()
}

// Back returns the last order in the queue without removing it.
func (oq *OrderQueue) Back() *Order {
	if oq.orders.Len() == 0 {
		return nil
	}
	return oq.orders.Back()
}

// PushBack adds an order to the back of the queue.
func (oq *OrderQueue) PushBack(order *Order) {
	oq.orders.PushBack(order)
}

// PopFront removes and returns the first order from the queue.
func (oq *OrderQueue) PopFront() *Order {
	if oq.orders.Len() == 0 {
		return nil
	}
	return oq.orders.PopFront()
}

// Remove removes a specific order from the queue.
func (oq *OrderQueue) Remove(order *Order) *Order {
	for i := 0; i < oq.orders.Len(); i++ {
		if oq.orders.At(i).GetID() == order.GetID() {
			return oq.orders.Remove(i)
		}
	}
	return nil
}