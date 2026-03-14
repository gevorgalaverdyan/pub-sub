package main

import (
	"fmt"
	"pubsub/db"
	"pubsub/publisher"
	"pubsub/validator"
	"sync"
)

func main() {
	db.ConnectDB()

	ue := make(chan publisher.UserEvent)
	validUe := make(chan publisher.UserEvent)

	ce := make(chan publisher.CommerceEvent)
	validCe := make(chan publisher.CommerceEvent)

	var ueWg sync.WaitGroup
	ueWg.Add(2)
	go func() { defer ueWg.Done(); publisher.StartPublishingUsers(ue) }()
	go func() { defer ueWg.Done(); publisher.StartPublishingGibberish(ue) }()
	go func() { ueWg.Wait(); close(ue) }()

	go publisher.StartPublishingCommerce(ce)

	go func() {
		defer close(validUe)
		for event := range ue {
			if err := validator.ValidateUserEvent(event); err != nil {
				fmt.Printf("❌ INVALID user event: %+v — %s\n", event, err)
				continue
			}
			validUe <- event
		}
	}()

	go func() {
		defer close(validCe)
		for event := range ce {
			if err := validator.ValidateCommerceEvent(event); err != nil {
				fmt.Printf("❌ INVALID commerce event: %+v — %s\n", event, err)
				continue
			}
			validCe <- event
		}
	}()

	go StartConsuming(validUe)
	go StartConsuming(validCe)

	select {}
}

func StartConsuming[T any](ch chan T) {
	for event := range ch {
		fmt.Printf("✅ Received: %+v\n", event)
	}
}
