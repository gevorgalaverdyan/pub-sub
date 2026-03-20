package publisher

import (
	"fmt"
	"math/rand"
	"pubsub/metrics"
	"time"
)

var pageEvents = []string{
	"page.viewed", "page.scrolled",  "order.placed", "payment.failed",
}

var paths = []string{"/home", "/about", "/pricing"}

type CommerceEvent struct {
	Event  string `json:"event"`
	UserId string `json:"user_id"`
	Path   string `json:"path"`
}

func StartPublishingCommerce(commQueue chan<- CommerceEvent, m *metrics.Metrics) {
	for range 100 {
		e := CommerceEvent{
			Event:  pageEvents[rand.Intn(len(pageEvents))],
			UserId: users[rand.Intn(len(users))],
			Path:   paths[rand.Intn(len(paths))],
		}

		commQueue <- e
		m.CommerceEvents.Inc()
		fmt.Printf("SENDING: %+v\n", e)

		time.Sleep(time.Duration(rand.Intn(4)) * time.Second)
	}
}
