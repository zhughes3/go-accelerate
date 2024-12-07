package pgx

import (
	"context"
	"errors"
	"github.com/georgysavva/scany/v2/pgxscan"

	"github.com/jackc/pgx/v5"
	"github.com/zhughes3/go-accelerate/pkg/slices"
)

type RowMapper[T any] func(pgx.Rows) (T, error)

func ExecInsertContextForPrimaryKey(ctx context.Context, query string, args ...any) (string, error) {
	var id string
	if err := mustGetContextTx(ctx).QueryRow(ctx, query, args...).Scan(&id); err != nil {
		return "", err
	}

	return id, nil
}

func ExecInsertContext(ctx context.Context, query string, args ...any) (bool, error) {
	rowCount, err := execContext(ctx, query, args...)
	return rowCount > 0, err
}

func ExecUpdateContext(ctx context.Context, query string, args ...any) (bool, error) {
	rowCount, err := execContext(ctx, query, args...)
	return rowCount > 0, err
}

func ExecDeleteContext(ctx context.Context, query string, args ...any) (bool, error) {
	rowCount, err := execContext(ctx, query, args...)
	return rowCount > 0, err
}

func RowsContext(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	return mustGetContextTx(ctx).Query(ctx, query, args...)
}

func QueryRowContext[T any](ctx context.Context, mapRow RowMapper[T], query string, args ...any) (T, bool, error) {
	entities, err := QueryContext(ctx, mapRow, query, args...)
	if err != nil || len(entities) == 0 {
		var zero T
		return zero, false, err
	}

	return entities[0], true, nil
}

func QueryContext[T any](ctx context.Context, mapRow RowMapper[T], query string, args ...any) ([]T, error) {
	rows, err := mustGetContextTx(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entities []T
	for rows.Next() {
		entity, err := mapRow(rows)
		if err != nil {
			return nil, err
		}

		entities = append(entities, entity)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return entities, nil
}

func ScanOneContext[T any](ctx context.Context, query string, args ...any) (T, bool, error) {
	var t T
	if err := pgxscan.Get(ctx, mustGetContextTx(ctx), &t, query, args...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = nil
		}

		return t, false, err
	}

	return t, true, nil
}

func ScanAndMapOneContext[T, S any](ctx context.Context, mapper func(T) S, query string, args ...any) (S, bool, error) {
	t, found, err := ScanOneContext[T](ctx, query, args...)
	if err != nil || !found {
		var zeroS S
		return zeroS, false, err
	}

	return mapper(t), true, nil
}

func ScanAllContext[T any](ctx context.Context, query string, args ...any) ([]T, error) {
	var t []T
	if err := pgxscan.Select(ctx, mustGetContextTx(ctx), &t, query, args...); err != nil {
		return nil, err
	}

	return t, nil
}

func ScanAndMapAllContext[T, S any](ctx context.Context, mapper func(T) S, query string, args ...any) ([]S, error) {
	ts, err := ScanAllContext[T](ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return slices.Map(ts, mapper), nil
}

func execContext(ctx context.Context, query string, args ...any) (int, error) {
	ct, err := mustGetContextTx(ctx).Exec(ctx, query, args...)
	if err != nil || ct.RowsAffected() == 0 {
		return 0, err
	}

	return int(ct.RowsAffected()), nil
}
