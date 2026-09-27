package domain

import "time"

type Notification struct {
	ID              string
	OrderID         string
	SourceEventID   string
	SourceEventType string
	Channel         string
	Recipient       string
	Subject         string
	Body            string
	Status          string
	DeliveredAt     time.Time
}
