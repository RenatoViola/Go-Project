package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

const (
	NUM_CHEFS   = 3
	NUM_WAITERS = 5
	RED         = "\033[31m"
	GREEN       = "\033[32m"
	YELLOW      = "\033[33m"
	CYAN        = "\033[36m"
	RESET       = "\033[0m"
)

type Order struct {
	id         int             // customer id
	dish       string          // dish description
	deliveryCh chan bool       // channel to transfer the order from waiter to customer
	ctxt       context.Context // customer context to apply a timeout
}

func Customer(id int, placedOrders chan<- Order) {

	// customer chooses something from the menu
	menuItem := [10]string{"Duck", "Veal", "Pork", "Salmon", "Venison",
		"Chicken", "Cod", "Tuna", "Swordfish", "Beast Burger"}[rand.Intn(10)]

	// create channel for customer to get the order from the waiter
	deliveryCh := make(chan bool)

	// derive a context from the background context, with the timeout correspondent to the customer's availability (3-5 seconds)
	ctxt, cancel := context.WithTimeout(context.Background(), (time.Duration(rand.Intn(3)+3) * time.Second))
	defer cancel()

	// place an order
	placedOrders <- Order{
		id:         id,
		dish:       menuItem,
		deliveryCh: deliveryCh,
		ctxt:       ctxt,
	}

	fmt.Printf("Customer %d arrived and ordered %s.\n",
		id, menuItem)

	select {
	case <-deliveryCh:
		// waiter will know that the customer received the dish successfully
	case <-ctxt.Done():
		fmt.Printf("%sCustomer %d waited for too long and left.%s\n", RED, id, RESET)
	}
}

func Chef(id int, placedOrders <-chan Order, readyOrders chan<- Order) {

	for order := range placedOrders {
		select {
		case <-order.ctxt.Done():
			fmt.Printf("%sChef %d won't prepare customer %d's %s, as the order was cancelled.%s\n", RED, id, order.id, order.dish, RESET)
			continue // skip this order and fetch the next order in line
		default:
			fmt.Printf("%sChef %d is preparing %s for customer %d.%s\n", CYAN, id, order.dish, order.id, RESET)

			time.Sleep(time.Duration(rand.Intn(3)+2) * time.Second) // chef takes 2 to 4 time units to prepare an order

			fmt.Printf("%sChef %d is done with customer %d's %s order.%s\n", YELLOW, id, order.id, order.dish, RESET)
			readyOrders <- order
		}
	}
}

func Waiter(id int, readyOrders <-chan Order) {

	for order := range readyOrders {
		select {
		case <-order.ctxt.Done():
			fmt.Printf("%sWaiter %d will discard customer %d's %s order, as it was cancelled.%s\n", RED, id, order.id, order.dish, RESET)
			continue // skip this order and fetch the next ready-to-deliver dish
		default:
			fmt.Printf("%sWaiter %d picked up the %s for customer %d.%s\n", GREEN, id, order.dish, order.id, RESET)
			// signal to the customer that the order arrived and block until the customer receives it (acknowledgment)
			order.deliveryCh <- true
			fmt.Printf("%sCustomer %d received the order of %s!%s\n", GREEN, order.id, order.dish, RESET)
		}
	}
}

func main() {

	placedOrders := make(chan Order, 10)
	readyOrders := make(chan Order, 5)

	for i := 1; i <= NUM_CHEFS; i++ {
		go Chef(i, placedOrders, readyOrders)
	}

	for i := 1; i <= NUM_WAITERS; i++ {
		go Waiter(i, readyOrders)
	}

	for i := 1; ; i++ {
		go Customer(i, placedOrders)
		time.Sleep(time.Duration(rand.Intn(i/5+1)) * time.Second)
	}
}
