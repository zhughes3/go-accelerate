package timelines

import (
	"context"
	"github.com/zhughes3/go-accelerate/internal/pkg/postgres"
	"github.com/zhughes3/go-accelerate/pkg/api"
)

func withTxService(db postgres.DB, service Service) Service {
	return txService{
		db:       db,
		delegate: service,
	}
}

type txService struct {
	db postgres.DB

	delegate Service
}

func (t txService) CreateTimeline(ctx context.Context, userID string, in TimelineCreateReq) (IdentifiableTimeline, error) {
	return postgres.ExecuteResultFuncInTx(ctx, t.db, func(ctx context.Context) (IdentifiableTimeline, error) {
		return t.delegate.CreateTimeline(ctx, userID, in)
	})
}

func (t txService) ListTimelines(ctx context.Context, userID string) (api.ListResponse[IdentifiableTimeline], error) {
	return postgres.ExecuteResultFuncInTx(ctx, t.db, func(ctx context.Context) (api.ListResponse[IdentifiableTimeline], error) {
		return t.delegate.ListTimelines(ctx, userID)
	})
}
