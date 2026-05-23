package database

import (
	"fmt"
	"os"

	"fcstask-monitor-bot/internal/logger"
	model "fcstask-monitor-bot/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to connect to database")
	}

	err = DB.AutoMigrate(&model.User{})
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to migrate user model")
	}

	logger.Log.Info().Msg("database initialized successfully")
}

func GetAllUsers() ([]model.User, error) {
	var users []model.User
	if err := DB.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
