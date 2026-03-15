package consumer

import (
	"pubsub/db"
	"pubsub/publisher"
)

const (
	USER_EVENT     = "USER_EVENT"
	COMMERCE_EVENT = "COMMERCE_EVENT"
)

func InsertUserEvent(ue publisher.UserEvent) error {
	var ueID string

	err := db.DB.QueryRow(`
		INSERT INTO events(event, user_id, timestamp, type) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id;`,
		ue.Event, ue.UserId, ue.Timestamp, USER_EVENT,
	).Scan(&ueID)

	if err != nil {
		return err
	}

	_, err = db.DB.Exec(`
		INSERT INTO properties (event_id, total, page)
		VALUES ($1, $2, $3)`, ueID, ue.Properties.Total, ue.Properties.Page,
	)

	return err
}

func InsertCommerceEvent(ce publisher.CommerceEvent) error {
	_, err := db.DB.Exec(`
		INSERT INTO events (event, user_id, path, type) VALUES ($1, $2, $3, $4)
	`, ce.Event, ce.UserId, ce.Path, COMMERCE_EVENT)

	return err
}