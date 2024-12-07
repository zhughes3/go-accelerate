package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBStats struct {
	MaxConnections          int
	CurrentConnections      int
	AvailableConnections    int
	AcquiredConnections     int
	EmptyAcquiredCount      int
	NewConnectionsTotal     int
	MaxLifetimeDestroyTotal int
	MaxIdleDestroyTotal     int
}

type BatchStatement struct {
	Query     string
	Arguments []any
}

type DB interface {
	Acquire(context.Context) (*pgxpool.Conn, error)
	Connect(context.Context) error
	Run(context.Context, func(*pgxpool.Conn) error) error
	Transaction(context.Context, func(pgx.Tx) error) error
	ContextWithTx(context.Context, pgx.Tx) context.Context
	BeginContextTx(context.Context) (context.Context, bool, error)
	BeginContextReadOnlyTx(context.Context) (context.Context, bool, error)
	CloseContextTx(context.Context, error) error
	NewBatch([]BatchStatement) *pgx.Batch
	Shutdown(context.Context) error
	SecurityString() [32]byte
	Stats() DBStats
}
