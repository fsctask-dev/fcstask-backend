package db

import (
	"context"
	"fmt"
	"time"

	"fcstask/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Client struct {
	db *gorm.DB
}

func New(cfg *config.DatabaseConfig) (*Client, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password,
		cfg.Database,
		cfg.SSLMode,
	)

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return &Client{db: db}, nil
}

func (c *Client) DB() *gorm.DB {
	return c.db
}

func (c *Client) Close() error {
	if c.db == nil {
		return nil
	}

	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

func (c *Client) AutoMigrate(models ...interface{}) error {
	return c.db.AutoMigrate(models...)
}

func (c *Client) WithContext(ctx context.Context) *gorm.DB {
	return c.db.WithContext(ctx)
}

func (c *Client) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return c.db.WithContext(ctx).Transaction(fn)
}
