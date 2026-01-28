package repo

import (
	"context"

	"gorm.io/gorm"

	"fcstask/internal/db/model"
)

type UserRepositoryInterface interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uint) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByUserID(ctx context.Context, userID int64) (*model.User, error)
	GetByTgUID(ctx context.Context, tgUID int64) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uint) error

	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	Count(ctx context.Context) (int64, error)
}

type UserRepository struct {
	db *gorm.DB
}

var _ UserRepositoryInterface = (*UserRepository)(nil)

func NewUserRepository(db *gorm.DB) UserRepositoryInterface {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByUserID(ctx context.Context, userID int64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByTgUID(ctx context.Context, tgUID int64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("tg_uid = ?", tgUID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, id).Error
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
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

func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
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

func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Count(&count).Error
	return count, err
}

func (r *UserRepository) ExistsByUserID(ctx context.Context, userID int64) (bool, error) {
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
func (r *UserRepository) ExistsByTgUID(ctx context.Context, tgUID int64) (bool, error) {
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
