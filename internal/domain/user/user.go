package user

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/zaynkorai/enlabs/internal/domain/transaction"
)

type User struct {
	ID        uint64          `json:"userId" gorm:"primaryKey"`
	Balance   decimal.Decimal `json:"balance" gorm:"type:numeric(20,2);default:0.00;not null"`
	CreatedAt time.Time       `gorm:"autoCreateTime"`
	UpdatedAt time.Time       `gorm:"autoUpdateTime"`
}

type Repository interface {
	GetByID(id uint64) (*User, error)
	AtomicUpdateBalanceAndCreateTransaction(userID uint64, newBalance decimal.Decimal, newTransaction *transaction.Transaction) error
	Create(user *User) error
}
