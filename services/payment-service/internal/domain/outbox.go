package domain

import "time"

type OutboxEvent struct {
	ID          string
	Topic       string
	EventKey    string
	EventType   string
	Payload     []byte
	CreatedAt   time.Time
	PublishedAt *time.Time
	Attempts    int
}
