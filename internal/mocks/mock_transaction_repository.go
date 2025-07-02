package mocks

import (
	"errors"

	"github.com/zaynkorai/enlabs/internal/domain/transaction"
)

type MockTransactionRepository struct {
	CreateFunc             func(transaction *transaction.Transaction) error
	GetByTransactionIDFunc func(transactionID string) (*transaction.Transaction, error)
}

func (m *MockTransactionRepository) Create(transaction *transaction.Transaction) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(transaction)
	}
	return errors.New("CreateFunc not set")
}

func (m *MockTransactionRepository) GetByTransactionID(transactionID string) (*transaction.Transaction, error) {
	if m.GetByTransactionIDFunc != nil {
		return m.GetByTransactionIDFunc(transactionID)
	}
	return nil, errors.New("GetByTransactionIDFunc not set")
}
