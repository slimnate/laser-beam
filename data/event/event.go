package event

import "time"

type Event struct {
	ID             int64
	Type           string
	Name           string
	Message        string
	Time           time.Time
	OrganizationID int64
}
