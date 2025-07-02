package transaction

import (
	"time"

	"github.com/shopspring/decimal"
)

type Transaction struct {
	ID            uint64          `json:"id" gorm:"primaryKey"`
	UserID        uint64          `json:"userId" gorm:"not null"`
	TransactionID string          `json:"transactionId" gorm:"unique;not null"` // External ID for idempotency
	SourceType    string          `json:"sourceType" gorm:"not null"`
	State         string          `json:"state" gorm:"not null"` // "win" or "lose"
	Amount        decimal.Decimal `json:"amount" gorm:"type:numeric(20,2);not null"`
	ProcessedAt   time.Time       `json:"processedAt" gorm:"autoCreateTime"`
}

type Repository interface {
	Create(transaction *Transaction) error
	GetByTransactionID(transactionID string) (*Transaction, error)
}
