package repo

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"fcstask-backend/internal/db"
	"fcstask-backend/internal/db/model"
)

type IUserRepo interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*model.User, error)
	GetByTgUID(ctx context.Context, tgUID int64) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetAllWithSessions(ctx context.Context, limit, offset int) ([]model.User, error)
	CountUsersWithSessions(ctx context.Context) (int64, error)

	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	Count(ctx context.Context) (int64, error)
}

type UserRepository struct {
	rw db.ReadWriter
}

var _ IUserRepo = (*UserRepository)(nil)

func NewUserRepository(rw db.ReadWriter) IUserRepo {
	return &UserRepository{rw: rw}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	return r.rw.WriteDB().WithContext(ctx).Create(user).Error
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.rw.ReadDB().WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.rw.ReadDB().WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.rw.ReadDB().WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.rw.ReadDB().WithContext(ctx).Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByTgUID(ctx context.Context, tgUID int64) (*model.User, error) {
	var user model.User
	err := r.rw.ReadDB().WithContext(ctx).Where("tg_uid = ?", tgUID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	return r.rw.WriteDB().WithContext(ctx).Save(user).Error
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.rw.WriteDB().WithContext(ctx).Delete(&model.User{}, "id = ?", id).Error
}

func (r *UserRepository) GetAllWithSessions(ctx context.Context, limit, offset int) ([]model.User, error) {
	readDB := r.rw.ReadDB()
	var users []model.User
	err := readDB.WithContext(ctx).
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
	readDB := r.rw.ReadDB()
	var count int64
	err := readDB.WithContext(ctx).
		Model(&model.User{}).
		Where("id IN (?)", readDB.Table("sessions").Select("DISTINCT user_id")).
		Count(&count).Error
	return count, err
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64

	err := r.rw.ReadDB().WithContext(ctx).
		Model(&model.User{}).
		Where("email = ?", email).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64

	err := r.rw.ReadDB().WithContext(ctx).
		Model(&model.User{}).
		Where("username = ?", username).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.rw.ReadDB().WithContext(ctx).Model(&model.User{}).Count(&count).Error
	return count, err
}

func (r *UserRepository) ExistsByUserID(ctx context.Context, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.rw.ReadDB().WithContext(ctx).
		Model(&model.User{}).
		Where("user_id = ?", userID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *UserRepository) ExistsByTgUID(ctx context.Context, tgUID int64) (bool, error) {
	var count int64
	err := r.rw.ReadDB().WithContext(ctx).
		Model(&model.User{}).
		Where("tg_uid = ?", tgUID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
