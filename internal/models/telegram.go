package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TelegramConfig struct {
	ID                    uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CompanyID             uuid.UUID `gorm:"type:uuid;uniqueIndex;notNull" json:"company_id"`
	APIKey                string    `gorm:"column:api_key;type:varchar(255);notNull" json:"api_key"`
	TelegramIDPenjualan   string    `gorm:"column:telegram_id_penjualan;type:varchar(50)" json:"telegram_id_penjualan"`
	TelegramIDPembelian   string    `gorm:"column:telegram_id_pembelian;type:varchar(50)" json:"telegram_id_pembelian"`
	TelegramIDStockOpname string    `gorm:"column:telegram_id_stock_opname;type:varchar(50)" json:"telegram_id_stock_opname"`
	NotifyPenjualan       bool      `gorm:"column:notify_penjualan;default:false" json:"notify_penjualan"`
	NotifyPembelian       bool      `gorm:"column:notify_pembelian;default:false" json:"notify_pembelian"`
	NotifyStockOpname     bool      `gorm:"column:notify_stock_opname;default:false" json:"notify_stock_opname"`
	IsActive              bool      `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt             time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (t *TelegramConfig) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (TelegramConfig) TableName() string {
	return "plan_telegram"
}
