package main

import (
	"fmt"
	"log"
	"net/http"
	"pubsub/consumer"
	"pubsub/db"
	"pubsub/metrics"
	"pubsub/publisher"
	"pubsub/validator"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	reg := prometheus.NewRegistry()
	m := metrics.NewMetrics(reg)

	db.ConnectDB()

	ue := make(chan publisher.UserEvent)
	validUe := make(chan publisher.UserEvent)

	ce := make(chan publisher.CommerceEvent)
	validCe := make(chan publisher.CommerceEvent)

	var ueWg sync.WaitGroup
	ueWg.Add(2)
	go func() { defer ueWg.Done(); publisher.StartPublishingUsers(ue, m) }()
	go func() { defer ueWg.Done(); publisher.StartPublishingGibberish(ue) }()
	go func() { ueWg.Wait(); close(ue) }()

	go publisher.StartPublishingCommerce(ce, m)

	go func() {
		defer close(validUe)
		for event := range ue {
			if err := validator.ValidateUserEvent(event); err != nil {
				m.ErrorsC.WithLabelValues("user", "validation").Inc()
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
				m.ErrorsC.WithLabelValues("commerce", "validation").Inc()
				fmt.Printf("❌ INVALID commerce event: %+v — %s\n", event, err)
				continue
			}
			validCe <- event
		}
	}()

	go StartConsuming(validUe, func(e publisher.UserEvent) error {
		return consumer.InsertUserEvent(e)
	}, m, "user")

	go StartConsuming(validCe, func(e publisher.CommerceEvent) error {
		return consumer.InsertCommerceEvent(e)
	}, m, "commerce")

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	log.Fatal(http.ListenAndServe(":8080", nil))
	select {}
}

func StartConsuming[T any](ch chan T, persist func(T) error, m *metrics.Metrics, stream string) {
	for event := range ch {
		if err := persist(event); err != nil {
			m.ErrorsC.WithLabelValues(stream, "persistance")
			log.Printf("FAILED persisting: %v\n", err)
			continue
		}

		fmt.Printf("✅ Received: %+v\n", event)
	}
}
