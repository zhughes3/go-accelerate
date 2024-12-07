package api

import "time"

type Event struct {
	Title       string    `json:"title,omitempty"`
	Timestamp   time.Time `json:"timestamp,omitempty"`
	Description string    `json:"description,omitempty"`
	Content     string    `json:"content,omitempty"`
	ImageURL    string    `json:"image_url,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

type IdentifiableEvent struct {
	ID         string `json:"id,omitempty"`
	TimelineID string `json:"timeline_id,omitempty"`
	Event
}

func NewIdentifiableEvent(id string, timelineID string, event Event) IdentifiableEvent {
	return IdentifiableEvent{
		ID:         id,
		TimelineID: timelineID,
		Event:      event,
	}
}

type Timeline struct {
	Name   string              `json:"name,omitempty"`
	Events []IdentifiableEvent `json:"events,omitempty"`
}

type IdentifiableTimeline struct {
	ID string `json:"id,omitempty"`
	Timeline
}

func NewIdentifiableTimeline(id string, timeline Timeline) IdentifiableTimeline {
	return IdentifiableTimeline{
		ID:       id,
		Timeline: timeline,
	}
}
