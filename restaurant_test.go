package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

/*
Successful delivery with correct sequence of events.
Order is placed, only then prepared and only then delivered.
Purposefully gave the customer a lot of patience to ensure a delivery
*/
func TestSuccessfulPreparationAndDelivery(t *testing.T) {
	placedOrders := make(chan Order, 10)
	readyOrders := make(chan Order, 5)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		Chef(1, placedOrders, readyOrders)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		Waiter(1, readyOrders)
	}()

	ctxt, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	deliveryCh := make(chan bool)
	order := Order{
		id:         1,
		dish:       "Lasagne",
		deliveryCh: deliveryCh,
		ctxt:       ctxt,
	}

	fmt.Printf("Customer %d arrived and ordered %s.\n", order.id, order.dish)
	placedOrders <- order

	select {
	case <-deliveryCh:

	case <-ctxt.Done():
		t.Errorf("Order was not delivered to the customer in time")
	}

	close(placedOrders)
	close(readyOrders)
	wg.Wait()
}
