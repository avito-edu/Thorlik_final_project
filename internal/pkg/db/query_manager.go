package db

import (
	"context"
	"database/sql"
	"fmt"

	"go.uber.org/zap"
)

type QueryManager struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewQueryManager(db *sql.DB, logger *zap.Logger) *QueryManager {
	return &QueryManager{
		db:     db,
		logger: logger,
	}
}

func (qm *QueryManager) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	qm.logger.Debug("Executing query row",
		zap.String("query", query))
	return qm.db.QueryRowContext(ctx, query, args...)
}

func (qm *QueryManager) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	qm.logger.Debug("Executing query",
		zap.String("query", query))
	rows, err := qm.db.QueryContext(ctx, query, args...)
	if err != nil {
		qm.logger.Error("Query execution failed",
			zap.Error(err),
			zap.String("query", query))
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	return rows, nil
}

func (qm *QueryManager) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	qm.logger.Debug("Executing statement",
		zap.String("query", query))
	result, err := qm.db.ExecContext(ctx, query, args...)
	if err != nil {
		qm.logger.Error("Statement execution failed",
			zap.Error(err),
			zap.String("query", query))
		return nil, fmt.Errorf("statement execution failed: %w", err)
	}
	return result, nil
}
