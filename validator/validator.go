package validator

import (
	"errors"
	"pubsub/publisher"
)

var validUserEvents = map[string]bool{
	"user.signed_up":  true,
	"feature.clicked": true,
}

var validCommerceEvents = map[string]bool{
	"page.viewed":    true,
	"page.scrolled":  true,
	"order.placed":   true,
	"payment.failed": true,
}

func ValidateUserEvent(e publisher.UserEvent) error {
	if e.Event == "" {
		return errors.New("event name is empty")
	}
	if !validUserEvents[e.Event] {
		return errors.New("unknown event: " + e.Event)
	}
	if e.UserId == "" {
		return errors.New("user_id is empty")
	}
	if e.Timestamp.IsZero() {
		return errors.New("timestamp is zero")
	}
	if e.Properties.Total < 0 {
		return errors.New("total is negative")
	}
	return nil
}

func ValidateCommerceEvent(e publisher.CommerceEvent) error {
	if e.Event == "" {
		return errors.New("event name is empty")
	}
	if !validCommerceEvents[e.Event] {
		return errors.New("unknown event: " + e.Event)
	}
	if e.UserId == "" {
		return errors.New("user_id is empty")
	}
	if e.Path == "" {
		return errors.New("path is empty")
	}
	return nil
}
