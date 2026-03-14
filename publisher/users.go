package publisher

import (
	"fmt"
	"math/rand"
	"time"
)

var userEvents = []string{
	"user.signed_up", "feature.clicked",
}

var users = []string{"user_001", "user_002", "user_003", "user_004", "user_005"}

type property struct {
	Total float64 `json:"total"`
	Page  string  `json:"page"`
}

type UserEvent struct {
	Event      string    `json:"event"`
	UserId     string    `json:"user_id"`
	Timestamp  time.Time `json:"timestamp"`
	Properties property  `json:"properties"`
}

func StartPublishingUsers(uQueue chan<- UserEvent) {
	for i := range 100 {
		e := UserEvent{
			Event:     userEvents[rand.Intn(len(userEvents))],
			UserId:    users[rand.Intn(len(users))],
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
			Properties: property{
				Total: rand.Float64() * 200,
				Page:  "/home",
			},
		}

		uQueue <- e
		fmt.Printf("SENDING: %+v\n", e)

		time.Sleep(time.Duration(rand.Intn(4)) * time.Second)
	}
}
