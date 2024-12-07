package pgx

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zhughes3/go-accelerate/internal/pkg/postgres"
	acerrors "github.com/zhughes3/go-accelerate/pkg/errors"
	"github.com/zhughes3/go-accelerate/pkg/slog"
	acsync "github.com/zhughes3/go-accelerate/pkg/sync"
	"sync"
	"time"
)

type db struct {
	logger slog.Logger

	sqldb  *pgxpool.Pool
	config *postgres.Config

	// state tracks the current state of the [pgxpool.Pool], "new", "started", or "shutdown"
	state *acsync.StateMachine

	// statsMu guards currentStats
	statsMu      sync.RWMutex
	currentStats postgres.DBStats

	// bgWG
	bgWG sync.WaitGroup
	// stop is a channel used to terminate background goroutines
	stop chan any
}

func NewDB(logger slog.Logger, config *postgres.Config) *db {
	_db := db{
		logger: logger,
		config: config,
		stop:   make(chan any),
		state: acsync.NewStateMachineBuilder(logger).
			WithComponentName("pgx_database").
			WithIgnoreAlreadyAtEndError(true).
			Build(),
	}

	return &_db
}

func NewDBConnect(ctx context.Context, logger slog.Logger, config *postgres.Config) (postgres.DB, error) {
	_db := NewDB(logger, config)
	return _db, _db.Connect(ctx)
}

func (db *db) Connect(ctx context.Context) error {
	return db.state.Start(ctx, func() error {
		db.logger.InfoContext(ctx, "Connecting to the database")
		if err := db.doConnect(ctx); err != nil {
			return err
		}
		db.logger.InfoContext(ctx, "Connecting to the database...complete")
		return nil
	})
}

func (db *db) doConnect(ctx context.Context) error {
	cc, err := db.config.ToPostgresConnConfig(db.logger)
	if err != nil {
		return err
	}

	cpc := db.config.NewPoolConfig(cc)

	for {
		if db.sqldb == nil {
			db.logger.InfoContext(ctx, "Attempting to connect to the database")
			if err := db.connect(ctx, cpc); err != nil {
				db.logger.WithError(err).ErrorContext(ctx, "Problem connecting to the database")
			}
			if db.sqldb == nil {
				db.logger.InfoContext(ctx, "Sleeping...")
				db.sleep()
				continue
			}
			db.logger.With("Database", db.config.Host).InfoContext(ctx, "Database connected")
		}
		break
	}

	// TODO collect stats

	return nil
}

func (db *db) connect(ctx context.Context, config *pgxpool.Config) error {
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return err
	}

	db.sqldb = pool

	return nil
}

func (db *db) sleep() {
	time.Sleep(time.Duration(db.config.ConnectionRetryWaitTime) * time.Second)
}

func (db *db) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	return db.doAcquire(ctx, db.config.ConnectionRetries)
}

func (db *db) doAcquire(ctx context.Context, retries int) (*pgxpool.Conn, error) {
	conn, err := db.sqldb.Acquire(ctx)
	if err != nil {
		if retries == 0 {
			return nil, acerrors.Wrap(err, "database unavailable")
		}

		db.logger.WithError(err).WarnContext(ctx, "Problem getting database connection; retry: %d", db.config.ConnectionRetries-retries+1)
		return db.doAcquire(ctx, retries-1)
	}

	return conn, nil
}

func (db *db) Run(ctx context.Context, f func(*pgxpool.Conn) error) error {
	//TODO implement me
	panic("implement me")
}

func (db *db) Transaction(ctx context.Context, f func(pgx.Tx) error) error {
	//TODO implement me
	panic("implement me")
}

func (db *db) ContextWithTx(ctx context.Context, tx pgx.Tx) context.Context {
	if _, ok := getContextTx(ctx); ok {
		return ctx
	}

	return context.WithValue(ctx, txContextKey, tx)
}

// BeginContextTx makes sure there is a default (read/write) transaction on the returned context.
// The bool return value indicates whether the caller should consider themselves to be the "owner" of the transaction.
// If true, the caller is responsible for closing the transaction.
func (db *db) BeginContextTx(ctx context.Context) (context.Context, bool, error) {
	return db.doBeginContextTx(ctx, pgx.TxOptions{})
}

func (db *db) doBeginContextTx(ctx context.Context, options pgx.TxOptions) (context.Context, bool, error) {
	if _, ok := getContextTx(ctx); ok {
		return ctx, false, nil
	}

	tx, err := db.sqldb.BeginTx(ctx, options)
	if err != nil {
		return nil, false, fmt.Errorf("could not begin tx: %w", err)
	}

	return context.WithValue(ctx, txContextKey, tx), true, nil
}

func (db *db) BeginContextReadOnlyTx(ctx context.Context) (context.Context, bool, error) {
	return db.doBeginContextTx(ctx, pgx.TxOptions{
		AccessMode: pgx.ReadOnly,
	})
}

func (db *db) CloseContextTx(ctx context.Context, err error) error {
	tx := mustGetContextTx(ctx)

	if err == nil {
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return fmt.Errorf("could not commit tx: %w", commitErr)
		}
	} else {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return fmt.Errorf("could not rollback tx: %s; original cause: %w", rollbackErr, err)
		}
	}

	return nil
}

func (db *db) NewBatch(statements []postgres.BatchStatement) *pgx.Batch {
	//TODO implement me
	panic("implement me")
}

func (db *db) Shutdown(ctx context.Context) error {
	return db.state.Shutdown(ctx, func() error {
		db.logger.InfoContext(ctx, "Shutting down the database...")

		// Tell all background goroutines to stop and terminate.
		close(db.stop)

		// Wait for all background goroutines to complete.
		db.bgWG.Wait()

		// Now that all background processes have been terminated, the database can be safely closed.
		db.sqldb.Close()
		db.logger.InfoContext(ctx, "Shutting down the database...complete")

		return nil
	})
}

func (db *db) SecurityString() [32]byte {
	return db.config.SecurityString
}

func (db *db) Stats() postgres.DBStats {
	//TODO implement me
	panic("implement me")
}

const txContextKey = contextKey("PGX_TX")

type contextKey string

func getContextTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txContextKey).(pgx.Tx)
	return tx, ok
}

func mustGetContextTx(ctx context.Context) pgx.Tx {
	tx, ok := getContextTx(ctx)
	if !ok {
		panic("no transaction on context")
	}

	return tx
}
