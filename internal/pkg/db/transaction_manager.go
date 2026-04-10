package db

import (
	"context"
	"database/sql"
	"fmt"

	"go.uber.org/zap"
)

type TransactionManager struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewTransactionManager(db *sql.DB, logger *zap.Logger) *TransactionManager {
	return &TransactionManager{
		db:     db,
		logger: logger,
	}
}

type Transaction struct {
	tx     *sql.Tx
	logger *zap.Logger
}

func (tm *TransactionManager) BeginTransaction(ctx context.Context) (*Transaction, error) {
	tm.logger.Debug("Starting transaction")
	tx, err := tm.db.BeginTx(ctx, nil)
	if err != nil {
		tm.logger.Error("Failed to start transaction", zap.Error(err))
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	return &Transaction{
		tx:     tx,
		logger: tm.logger,
	}, nil
}

func (t *Transaction) Commit() error {
	t.logger.Debug("Committing transaction")
	if err := t.tx.Commit(); err != nil {
		t.logger.Error("Failed to commit transaction", zap.Error(err))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (t *Transaction) Rollback() error {
	t.logger.Debug("Rolling back transaction")
	if err := t.tx.Rollback(); err != nil {
		if err != sql.ErrTxDone {
			t.logger.Error("Failed to rollback transaction", zap.Error(err))
			return fmt.Errorf("failed to rollback transaction: %w", err)
		}
	}
	return nil
}

func (t *Transaction) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	t.logger.Debug("Executing query row in transaction",
		zap.String("query", query))
	return t.tx.QueryRowContext(ctx, query, args...)
}

func (t *Transaction) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	t.logger.Debug("Executing query in transaction",
		zap.String("query", query))
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		t.logger.Error("Query execution failed in transaction",
			zap.Error(err),
			zap.String("query", query))
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	return rows, nil
}

func (t *Transaction) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	t.logger.Debug("Executing statement in transaction",
		zap.String("query", query))
	result, err := t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		t.logger.Error("Statement execution failed in transaction",
			zap.Error(err),
			zap.String("query", query))
		return nil, fmt.Errorf("statement execution failed: %w", err)
	}
	return result, nil
}

func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(*Transaction) error) error {
	tx, err := tm.BeginTransaction(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			tm.logger.Error("Failed to rollback transaction after error",
				zap.Error(rbErr),
				zap.Error(err))
		}
		return err
	}

	return tx.Commit()
}
