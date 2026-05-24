package main

import (
    "database/sql"
    "fmt"
    "log"
    "path/filepath"

    "fcstask-backend/internal/config"

    _ "github.com/lib/pq"
    "github.com/pressly/goose/v3"
)

func main() {
    cfg, err := config.Load("config/config.yaml")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    dsn := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        cfg.Database.Host,
        cfg.Database.Port,
        cfg.Database.Username,
        cfg.Database.Password,
        cfg.Database.Database,
        cfg.Database.SSLMode,
    )

    db, err := sql.Open("postgres", dsn)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        log.Fatalf("Database is not responding: %v", err)
    }

    migrationsDir := filepath.Join("internal", "db", "migration")
    goose.SetDialect("postgres")
    log.Printf("Running migrations from: %s", migrationsDir)

    if err := goose.Up(db, migrationsDir); err != nil {
        log.Fatalf("Migration failed: %v", err)
    }

    log.Println("Migrations applied successfully")
}