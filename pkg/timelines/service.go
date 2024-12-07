package timelines

import (
	"context"
	"github.com/zhughes3/go-accelerate/internal/pkg/postgres"
	"github.com/zhughes3/go-accelerate/internal/pkg/v1/timelines"
	"github.com/zhughes3/go-accelerate/pkg/api"
	"github.com/zhughes3/go-accelerate/pkg/slog"
	"time"
)

type Service interface { // CreateTimeline validates and stores the given timeline.
	CreateTimeline(ctx context.Context, userID string, in TimelineCreateReq) (IdentifiableTimeline, error)

	// ListTimelines grabs all timelines associated with the user.
	ListTimelines(ctx context.Context, userID string) (api.ListResponse[IdentifiableTimeline], error)
}

type service struct {
	logger slog.Logger
	dao    timelines.DAO
}

func NewService(logger slog.Logger, db postgres.DB) Service {
	return withTxService(db, service{
		logger: logger,
		dao:    timelines.NewDAO(logger),
	})
}

func (s service) CreateTimeline(ctx context.Context, userID string, in TimelineCreateReq) (IdentifiableTimeline, error) {
	now := time.Now()
	timelineID, err := s.dao.CreateTimeline(ctx, userID, timelines.TimelineCreateRequest{
		Name:      in.Name,
		CreatedAt: now,
	})
	if err != nil {
		return IdentifiableTimeline{}, err
	}

	return IdentifiableTimeline{
		ID: timelineID,
		Timeline: Timeline{
			Name:      in.Name,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}, nil
}

func (s service) ListTimelines(ctx context.Context, userID string) (api.ListResponse[IdentifiableTimeline], error) {
	resp, err := s.dao.ListTimelines(ctx, userID)
	if err != nil {
		return api.NewListResponse([]IdentifiableTimeline{}), err
	}

	return api.NewListResponse(toIdentifiableTimelines(resp)), nil
}

func toIdentifiableTimelines(ts []timelines.Timeline) []IdentifiableTimeline {
	var resp []IdentifiableTimeline
	for _, t := range ts {
		resp = append(resp, IdentifiableTimeline{
			ID:     t.ID,
			UserID: t.UserID,
			Timeline: Timeline{
				Name:      t.Name,
				CreatedAt: t.CreatedAt,
				UpdatedAt: t.UpdatedAt,
			},
		})
	}

	return resp
}

type IdentifiableTimeline struct {
	ID     string
	UserID string
	Timeline
}

type Timeline struct {
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TimelineCreateReq struct {
	Name string
}
