package database

import (
	"context"

	"goplace_backend/internal/model"

	"gorm.io/gorm"
)

type UserDatabase struct {
	db *gorm.DB
}

func NewUserDatabase(db *gorm.DB) *UserDatabase {
	return &UserDatabase{db: db}
}

func (d *UserDatabase) Create(ctx context.Context, user *model.User) (*model.User, error) {
	result := d.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}

func (d *UserDatabase) Update(ctx context.Context, user *model.User) (*model.User, error) {
	result := d.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}

func (d *UserDatabase) GetById(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	result := d.db.WithContext(ctx).First(&user, id)
	return &user, result.Error
}

func (d *UserDatabase) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	result := d.db.WithContext(ctx).Where("username = ?", username).Limit(1).Find(&user)
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &user, result.Error
}

func (d *UserDatabase) GetLeaderboard(ctx context.Context, page, size int) ([]model.User, error) {
	var users []model.User
	result := d.db.WithContext(ctx).
		Order("amount_placed DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&users)
	return users, result.Error
}
