package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type ConnectionManager struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewConnectionManager(cfg *DatabaseConfig, logger *zap.Logger) (*ConnectionManager, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established",
		zap.String("host", cfg.Host),
		zap.String("database", cfg.DBName))

	return &ConnectionManager{
		db:     db,
		logger: logger,
	}, nil
}

func (cm *ConnectionManager) GetDB() *sql.DB {
	return cm.db
}

func (cm *ConnectionManager) Close() error {
	cm.logger.Info("Closing database connection")
	return cm.db.Close()
}

func (cm *ConnectionManager) HealthCheck(ctx context.Context) error {
	return cm.db.PingContext(ctx)
}
