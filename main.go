package main

import (
	"fmt"
	"log"
	"pubsub/consumer"
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

	go StartConsuming(validUe, func(e publisher.UserEvent) error {
		return consumer.InsertUserEvent(e)
	})

	go StartConsuming(validCe, func(e publisher.CommerceEvent) error {
		return consumer.InsertCommerceEvent(e)
	})

	select {}
}

func StartConsuming[T any](ch chan T, persist func(T) error) {
	for event := range ch {
		if err := persist(event); err != nil {
			log.Printf("FAILED persisting: %v\n", err)
			continue
		}

		fmt.Printf("✅ Received: %+v\n", event)
	}
}
