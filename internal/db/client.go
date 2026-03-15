package db

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"fcstask/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ReadWriter interface {
	ReadDB() *gorm.DB
	WriteDB() *gorm.DB
}

type Client struct {
	master   *gorm.DB
	replicas []*gorm.DB
	readIdx  atomic.Uint64
}

func openDB(dsn string, gormConfig *gorm.Config, maxOpen, maxIdle int, connMaxLifetime time.Duration) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		_ = db
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, err
	}
	return db, nil
}

func dsn(host string, port int, username, password, database, sslMode string) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, username, password, database, sslMode,
	)
}

func New(cfg *config.DatabaseConfig) (*Client, error) {
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	master, err := openDB(dsn(cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, cfg.SSLMode), gormConfig, cfg.MaxOpenConns, cfg.MaxIdleConns, cfg.ConnMaxLifetime)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to master: %w", err)
	}

	client := &Client{master: master}

	for i, r := range cfg.Replicas {
		replica, err := openDB(dsn(r.Host, r.Port, r.Username, r.Password, r.Database, r.SSLMode), gormConfig, cfg.MaxOpenConns, cfg.MaxIdleConns, cfg.ConnMaxLifetime)
		if err != nil {
			client.closeAll()
			return nil, fmt.Errorf("failed to connect to replica %d (%s:%d): %w", i+1, r.Host, r.Port, err)
		}
		client.replicas = append(client.replicas, replica)
	}

	return client, nil
}

func (c *Client) ReadDB() *gorm.DB {
	if n := len(c.replicas); n > 0 { // round-robin
		idx := c.readIdx.Add(1) - 1
		return c.replicas[idx%uint64(n)]
	}
	return c.master
}

func (c *Client) WriteDB() *gorm.DB {
	return c.master
}

// DB returns master. Use ReadDB/WriteDB for read/write split.
func (c *Client) DB() *gorm.DB {
	return c.master
}

func (c *Client) closeAll() {
	_ = c.Close()
}

func (c *Client) Close() error {
	var err error
	if c.master != nil {
		if sqlDB, e := c.master.DB(); e == nil {
			err = sqlDB.Close()
		}
		c.master = nil
	}
	for _, replica := range c.replicas {
		if replica != nil {
			if sqlDB, e := replica.DB(); e == nil {
				if e2 := sqlDB.Close(); e2 != nil {
					err = e2
				}
			}
		}
	}
	c.replicas = nil
	return err
}

func (c *Client) AutoMigrate(models ...interface{}) error {
	return c.master.AutoMigrate(models...)
}

func (c *Client) WithContext(ctx context.Context) *gorm.DB {
	return c.master.WithContext(ctx)
}

func (c *Client) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return c.master.WithContext(ctx).Transaction(fn)
}
