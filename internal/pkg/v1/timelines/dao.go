package timelines

import (
	"context"
	_ "embed"
	"github.com/zhughes3/go-accelerate/internal/pkg/postgres/pgx"
	acerrors "github.com/zhughes3/go-accelerate/pkg/errors"
	"github.com/zhughes3/go-accelerate/pkg/slog"
)

type dao struct {
	logger slog.Logger
}

type DAO interface {
	CreateTimeline(ctx context.Context, userID string, req TimelineCreateRequest) (string, error)
	ListTimelines(ctx context.Context, userID string) ([]Timeline, error)
}

func NewDAO(logger slog.Logger) DAO {
	return dao{logger}
}

var (
	//go:embed sql/insert-timeline.sql
	insertTimeline string

	//go:embed sql/list-timelines.sql
	listTimelines string
)

func (d dao) CreateTimeline(ctx context.Context, userID string, req TimelineCreateRequest) (string, error) {
	timelineID, err := pgx.ExecInsertContextForPrimaryKey(ctx, insertTimeline, req.Name, userID, req.CreatedAt, req.CreatedAt)
	if err != nil {
		return "", acerrors.Wrap(err, "problem inserting timeline")
	}

	return timelineID, nil
}

func (d dao) ListTimelines(ctx context.Context, userID string) ([]Timeline, error) {
	return pgx.ScanAllContext[Timeline](ctx, listTimelines, userID)
}
