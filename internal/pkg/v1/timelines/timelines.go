package timelines

import "time"

type Timeline struct {
	ID        string
	UserID    string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TimelineCreateRequest struct {
	Name      string
	CreatedAt time.Time
}
