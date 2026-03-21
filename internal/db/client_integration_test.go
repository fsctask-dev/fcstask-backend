//go:build integration

package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"fcstask-backend/internal/config"
	"fcstask-backend/internal/db/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func integrationConfig(t *testing.T) *config.DatabaseConfig {
	t.Helper()
	host := getenv("PG_HOST", "localhost")
	masterPort := atoi(t, getenv("PG_MASTER_PORT", "6432"))
	replicaPort := atoi(t, getenv("PG_REPLICA_PORT", "6433"))
	return &config.DatabaseConfig{
		Host:            host,
		Port:            masterPort,
		Username:        "postgres",
		Password:        "postgres",
		Database:        "fcstask",
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Minute,
		Replicas: []config.ReplicaConfig{
			{
				Host:     host,
				Port:     replicaPort,
				Username: "postgres",
				Password: "postgres",
				Database: "fcstask",
				SSLMode:  "disable",
			},
		},
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func atoi(t *testing.T, s string) int {
	t.Helper()
	n, err := strconv.Atoi(s)
	if err != nil {
		t.Fatalf("invalid int %q: %v", s, err)
	}
	return n
}

func migrateUserAndSession(t *testing.T, client *Client) error {
	t.Helper()
	tx := client.WriteDB().Session(&gorm.Session{})
	tx.Config.IgnoreRelationshipsWhenMigrating = true
	if err := tx.AutoMigrate(&model.User{}); err != nil {
		return fmt.Errorf("AutoMigrate User: %w", err)
	}
	if err := tx.AutoMigrate(&model.Session{}); err != nil {
		return fmt.Errorf("AutoMigrate Session: %w", err)
	}
	return nil
}

func TestClientReplicationReadWriteSplit(t *testing.T) {
	ctx := context.Background()
	cfg := integrationConfig(t)

	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	if err := migrateUserAndSession(t, client); err != nil {
		t.Fatal(err)
	}

	uid := uuid.New()
	email := fmt.Sprintf("repl-test-%s@example.com", uid.String()[:8])
	u := &model.User{
		Email:        email,
		Username:     "u_" + uid.String()[:8],
		PasswordHash: "x",
		UserID:       uid,
	}
	if err := client.WriteDB().WithContext(ctx).Create(u).Error; err != nil {
		t.Fatalf("write Create: %v", err)
	}

	var fromMaster model.User
	if err := client.ReadMasterDB().WithContext(ctx).Where("email = ?", email).First(&fromMaster).Error; err != nil {
		t.Fatalf("ReadMasterDB First: %v", err)
	}
	if fromMaster.ID != u.ID {
		t.Fatalf("ReadMasterDB: got id %v want %v", fromMaster.ID, u.ID)
	}

	deadline := time.Now().Add(30 * time.Second)
	var fromReplica model.User
	var repErr error
	for time.Now().Before(deadline) {
		repErr = client.ReadDB().WithContext(ctx).Where("email = ?", email).First(&fromReplica).Error
		if repErr == nil {
			break
		}
		if !errors.Is(repErr, gorm.ErrRecordNotFound) {
			t.Fatalf("ReadDB First: %v", repErr)
		}
		time.Sleep(200 * time.Millisecond)
	}
	if repErr != nil {
		t.Fatalf("ReadDB: row not visible on replica after timeout: %v", repErr)
	}
	if fromReplica.ID != u.ID {
		t.Fatalf("ReadDB: got id %v want %v", fromReplica.ID, u.ID)
	}
}

func TestClientTransaction(t *testing.T) {
	ctx := context.Background()
	cfg := integrationConfig(t)
	cfg.Replicas = nil

	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	if err := migrateUserAndSession(t, client); err != nil {
		t.Fatal(err)
	}

	uid := uuid.New()
	email := fmt.Sprintf("tx-test-%s@example.com", uid.String()[:8])
	u := &model.User{
		Email:        email,
		Username:     "tx_" + uid.String()[:8],
		PasswordHash: "x",
		UserID:       uid,
	}

	err = client.Transaction(ctx, func(tx *gorm.DB) error {
		return tx.Create(u).Error
	})
	if err != nil {
		t.Fatalf("Transaction: %v", err)
	}

	var got model.User
	if err := client.ReadMasterDB().WithContext(ctx).Where("email = ?", email).First(&got).Error; err != nil {
		t.Fatalf("after tx First: %v", err)
	}
	if got.ID != u.ID {
		t.Fatalf("id mismatch")
	}
}

var _ ReadWriter = (*Client)(nil)
