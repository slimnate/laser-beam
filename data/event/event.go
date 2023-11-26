package event

import "time"

type Event struct {
	ID             int64
	Type           string
	Application    string
	Name           string
	Message        string
	Time           time.Time
	OrganizationID int64
}

func (e *Event) FormattedTime() string {
	return e.Time.Format("2006/01/02 15:04:05")
}
