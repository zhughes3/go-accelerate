package postgres

import "context"

// A ResultFunc allows an operation that takes a context and
// returns a value and an error to be wrapped in a transaction.
type ResultFunc[T any] func(ctx context.Context) (T, error)

// A ResultAndFoundFunc allows an operation that takes a context and
// returns a value, a found indicator, and an error to be wrapped in a tx.
type ResultAndFoundFunc[T any] func(ctx context.Context) (T, bool, error)

type beginTxFunc func(ctx context.Context) (context.Context, bool, error)

func ExecuteResultFuncInTx[T any](ctx context.Context, db DB, op ResultFunc[T]) (T, error) {
	return doExecuteResultFuncInTx[T](ctx, db, db.BeginContextTx, op)
}

func doExecuteResultFuncInTx[T any](ctx context.Context, db DB, beginTx beginTxFunc, op ResultFunc[T]) (T, error) {
	var zero T

	ctx, owned, err := beginTx(ctx)
	if err != nil {
		return zero, err
	}

	result, err := op(ctx)

	if owned {
		if closeErr := db.CloseContextTx(ctx, err); closeErr != nil {
			return zero, closeErr
		}
	}

	return result, err
}
