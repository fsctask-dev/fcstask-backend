package repo

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fcstask-backend/internal/db/model"
)

type IUserRepo interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	GetUserByUserID(ctx context.Context, userID uuid.UUID) (*model.User, error)
	GetUserByTgUID(ctx context.Context, tgUID int64) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error

	GetUsersWithSessions(ctx context.Context, limit, offset int) ([]model.User, error)
	CountUsersWithSessions(ctx context.Context) (int64, error)

	ExistsUserByEmail(ctx context.Context, email string) (bool, error)
	ExistsUserByUsername(ctx context.Context, username string) (bool, error)
	CountUsers(ctx context.Context) (int64, error)
}

type UserRepository struct {
	db *gorm.DB
}

var _ IUserRepo = (*UserRepository)(nil)

func NewUserRepository(db *gorm.DB) IUserRepo {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByUserID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByTgUID(ctx context.Context, tgUID int64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("tg_uid = ?", tgUID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *UserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, "id = ?", id).Error
}

func (r *UserRepository) GetUsersWithSessions(ctx context.Context, limit, offset int) ([]model.User, error) {
	var users []model.User
	err := r.db.WithContext(ctx).
		Joins("JOIN sessions ON sessions.user_id = users.id").
		Group("users.id").
		Preload("Sessions", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Order("users.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) CountUsersWithSessions(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id IN (?)", r.db.Table("sessions").Select("DISTINCT user_id")).
		Count(&count).Error
	return count, err
}

func (r *UserRepository) ExistsUserByEmail(ctx context.Context, email string) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("email = ?", email).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *UserRepository) ExistsUserByUsername(ctx context.Context, username string) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("username = ?", username).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *UserRepository) CountUsers(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Count(&count).Error
	return count, err
}

func (r *UserRepository) ExistsUserByUserID(ctx context.Context, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("user_id = ?", userID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsByTgUID проверяет существование пользователя по tg_uid
func (r *UserRepository) ExistsUserByTgUID(ctx context.Context, tgUID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("tg_uid = ?", tgUID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
