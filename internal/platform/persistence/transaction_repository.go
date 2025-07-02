package persistence

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/zaynkorai/enlabs/internal/domain/transaction"
	"gorm.io/gorm"
)

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(t *transaction.Transaction) error {
	if err := r.db.Create(t).Error; err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}
	return nil
}

func (r *TransactionRepository) GetByTransactionID(transactionID string) (*transaction.Transaction, error) {
	var t transaction.Transaction
	result := r.db.Where("transaction_id = ?", transactionID).First(&t)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get transaction by ID %s: %w", transactionID, result.Error)
	}
	return &t, nil
}
