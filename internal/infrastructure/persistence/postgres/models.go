package postgres

import (
	"time"
)

// RateModel represents the database table for exchange rates.
type RateModel struct {
	ID            string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	BaseCurrency  string    `gorm:"type:varchar(3);not null;index:idx_pair_date"`
	QuoteCurrency string    `gorm:"type:varchar(3);not null;index:idx_pair_date"`
	Value         float64   `gorm:"type:decimal(20,10);not null"`
	EffectiveDate time.Time `gorm:"type:date;not null;index:idx_pair_date"`
	Source        string    `gorm:"type:varchar(50);not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// TableName specifies the table name for RateModel.
func (RateModel) TableName() string {
	return "exchange_rates"
}
