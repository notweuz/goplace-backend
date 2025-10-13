package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username     string `gorm:"unique"`
	Password     []byte
	LastPlaced   time.Time
	Active       bool `gorm:"default:true"`
	Admin        bool `gorm:"not null,default:false"`
	Banned       bool `gorm:"not null;default:false"`
	PixelsStock  uint `gorm:"default:1"`
	AmountPlaced int  `gorm:"default:0"`
	TokenVersion uint `gorm:"default:0"`
}

func NewUser(username, password string) *User {
	return &User{
		Username:     username,
		Password:     []byte(password),
		Active:       true,
		Admin:        false,
		Banned:       false,
		LastPlaced:   time.Now(),
		PixelsStock:  1,
		AmountPlaced: 0,
		TokenVersion: 0,
	}
}
