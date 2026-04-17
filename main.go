package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pubsub/consumer"
	"pubsub/db"
	"pubsub/metrics"
	"pubsub/publisher"
	"pubsub/validator"
	"sync"
	"time"

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

	mux := http.NewServeMux()

	mux.Handle("/prune", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := db.PruneDB(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte(err.Error()))
			if err != nil {
				log.Printf("Failed to write response: %v", err)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("DB Nuked"))
		if err != nil {
			log.Printf("Failed to write response: %v", err)
		}
	}))

	// GET /events – returns the 50 most recent stored events as JSON.
	mux.Handle("/events", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		events, err := db.QueryEvents(50)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Failed to query events: %v", err)
			return
		}
		if events == nil {
			events = []db.Event{}
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"events": events}); err != nil {
			log.Printf("Failed to encode events: %v", err)
		}
	}))

	// POST /publish – validates and persists a manually submitted event.
	mux.Handle("/publish", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Type   string  `json:"type"`
			Event  string  `json:"event"`
			UserID string  `json:"user_id"`
			Total  float64 `json:"total"`
			Page   string  `json:"page"`
			Path   string  `json:"path"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON: " + err.Error()}) //nolint:errcheck
			return
		}

		w.Header().Set("Content-Type", "application/json")

		switch req.Type {
		case "user":
			e := publisher.UserEvent{
				Event:     req.Event,
				UserId:    req.UserID,
				Timestamp: time.Now(),
			}
			e.Properties.Total = req.Total
			e.Properties.Page = req.Page

			if err := validator.ValidateUserEvent(e); err != nil {
				m.ErrorsC.WithLabelValues("user", "validation").Inc()
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error(), "status": "invalid"}) //nolint:errcheck
				return
			}
			if err := consumer.InsertUserEvent(e); err != nil {
				m.ErrorsC.WithLabelValues("user", "persistance").Inc()
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}) //nolint:errcheck
				return
			}
			m.UserEvents.Inc()

		case "commerce":
			e := publisher.CommerceEvent{
				Event:  req.Event,
				UserId: req.UserID,
				Path:   req.Path,
			}

			if err := validator.ValidateCommerceEvent(e); err != nil {
				m.ErrorsC.WithLabelValues("commerce", "validation").Inc()
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error(), "status": "invalid"}) //nolint:errcheck
				return
			}
			if err := consumer.InsertCommerceEvent(e); err != nil {
				m.ErrorsC.WithLabelValues("commerce", "persistance").Inc()
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}) //nolint:errcheck
				return
			}
			m.CommerceEvents.Inc()

		default:
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "unknown type: use 'user' or 'commerce'"}) //nolint:errcheck
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"status": "published"}) //nolint:errcheck
	}))

	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	// Serve the frontend. Must be registered last – "/" is a catch-all.
	mux.Handle("/", http.FileServer(http.Dir("./frontend")))

	log.Fatal(http.ListenAndServe(":8080", mux))
}

func StartConsuming[T any](ch chan T, persist func(T) error, m *metrics.Metrics, stream string) {
	for event := range ch {
		if err := persist(event); err != nil {
			m.ErrorsC.WithLabelValues(stream, "persistance").Inc()
			log.Printf("FAILED persisting: %v\n", err)
			continue
		}

		fmt.Printf("✅ Received: %+v\n", event)
	}
}
