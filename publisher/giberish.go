package publisher

import (
	"math/rand"
	"time"
)

func StartPublishingGibberish(ch chan<- UserEvent) {
    gibberishEvents := []string{
        "", "???", "12345", "not_an_event", "DROP TABLE events",
    }
    gibberishUsers := []string{
        "", "not-a-uuid", "12345", "null", "undefined",
    }

    for range 50 {
        ch <- UserEvent{
            Event:     gibberishEvents[rand.Intn(len(gibberishEvents))],
            UserId:    gibberishUsers[rand.Intn(len(gibberishUsers))],
            Timestamp: time.Time{},
            Properties: property{
                Total: -999.99, 
                Page:  "",      
            },
        }
        time.Sleep(10 * time.Second)
    }
}