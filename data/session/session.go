package session

import (
	"time"
)

type Session struct {
	ID        int64
	Key       string
	UserID    int64
	LoginTime time.Time
	LastSeen  time.Time
}
