package model

import (
	accountModel "kek-backend/internal/account/model"
	"time"
)

type Alert struct {
	ID             uint      `gorm:"column:id"`
	Slug           string    `gorm:"column:slug"`
	Title          string    `gorm:"column:title"`
	Body           string    `gorm:"column:body"`
	PairAddress    string    `gorm:"column:pair_address"`
	AlertType      string    `gorm:"column:alert_type"`
	AlertValue     string    `gorm:"column:alert_value"`
	AlertOption    string    `gorm:"column:alert_option"`
	ExpirationTime time.Time `gorm:"column:expiration_time"`
	AlertActions   []string  `gorm:"column:alert_actions"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
	DeletedAtUnix  int64     `gorm:"column:deleted_at_unix"`
	Author         accountModel.Account
	AuthorID       uint
}
