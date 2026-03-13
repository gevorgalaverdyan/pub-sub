package main

import "pubsub/publisher"

func main() {
	e := []publisher.Event{}
	publisher.StartPublishingUsers(e)
	
}