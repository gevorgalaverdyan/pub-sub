package publisher

import (
	"fmt"
	"math/rand"
	"time"
)

var events = []string{
	"user.signed_up", "page.viewed", "order.placed",
	"feature.clicked", "payment.failed",
}

var users = []string{"user_001", "user_002", "user_003", "user_004", "user_005"}

type Property struct {
	Total float64 `json:"total"`
	Page string `json:"page"`
}

type Event struct {
	Event string `json:"event"`
	UserId string `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
	Properties Property `json:"properties"`
}

func StartPublishingUsers(userEvents []Event) {
	for i:=range 100 {
		userEvents = append(userEvents, 
			Event{
				Event: events[rand.Intn(len(events))],
				UserId: users[rand.Intn(len(users))],
				Timestamp: time.Now().Add(time.Duration(i)),
				Properties: Property{
					Total: rand.Float64() * 200,
					Page: "/home",
				},
			},
		)
		time.Sleep(time.Duration(rand.Intn(4)) * time.Second)
		fmt.Printf("Published %+v\n", userEvents[len(userEvents)-1])
	}
}
