package database

import (
	"context"

	"goplace_backend/internal/model"

	"gorm.io/gorm"
)

type PixelDatabase struct {
	db *gorm.DB
}

func NewPixelDatabase(db *gorm.DB) *PixelDatabase {
	return &PixelDatabase{db: db}
}

func (d *PixelDatabase) Create(ctx context.Context, pixel *model.Pixel) (*model.Pixel, error) {
	result := d.db.WithContext(ctx).Create(pixel)
	if result.Error != nil {
		return nil, result.Error
	}

	err := d.db.WithContext(ctx).Preload("User").First(pixel, pixel.ID).Error
	if err != nil {
		return nil, err
	}

	return pixel, nil
}

func (d *PixelDatabase) Update(ctx context.Context, pixel *model.Pixel) (*model.Pixel, error) {
	result := d.db.WithContext(ctx).Save(pixel)
	if result.Error != nil {
		return nil, result.Error
	}

	err := d.db.WithContext(ctx).Preload("User").First(pixel, pixel.ID).Error
	if err != nil {
		return nil, err
	}

	return pixel, nil
}

func (d *PixelDatabase) GetByID(ctx context.Context, id uint) (*model.Pixel, error) {
	var pixel model.Pixel
	result := d.db.WithContext(ctx).Preload("User").First(&pixel, id)
	return &pixel, result.Error
}

func (d *PixelDatabase) GetAll(ctx context.Context) ([]model.Pixel, error) {
	var pixels []model.Pixel
	return pixels, d.db.WithContext(ctx).Preload("User").Find(&pixels).Error
}

func (d *PixelDatabase) GetByCoordinates(ctx context.Context, x, y uint) (*model.Pixel, error) {
	var pixel model.Pixel
	err := d.db.WithContext(ctx).Preload("User").Where("x = ? AND y = ?", x, y).First(&pixel).Error
	return &pixel, err
}

func (d *PixelDatabase) GetAllByUserID(ctx context.Context, userId uint) ([]model.Pixel, error) {
	var pixels []model.Pixel
	return pixels, d.db.WithContext(ctx).Preload("User").Where("user_id = ?", userId).Find(&pixels).Error
}

func (d *PixelDatabase) Delete(ctx context.Context, id uint) error {
	result := d.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Pixel{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
