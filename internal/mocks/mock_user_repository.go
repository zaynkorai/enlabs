package mocks

import (
	"errors"

	"github.com/shopspring/decimal"
	"github.com/zaynkorai/enlabs/internal/domain/transaction"
	"github.com/zaynkorai/enlabs/internal/domain/user"
)

type MockUserRepository struct {
	GetByIDFunc                                 func(id uint64) (*user.User, error)
	AtomicUpdateBalanceAndCreateTransactionFunc func(userID uint64, newBalance decimal.Decimal, newTransaction *transaction.Transaction) error
	CreateFunc                                  func(user *user.User) error
}

func (m *MockUserRepository) GetByID(id uint64) (*user.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, errors.New("GetByIDFunc not set")
}

func (m *MockUserRepository) AtomicUpdateBalanceAndCreateTransaction(userID uint64, newBalance decimal.Decimal, newTransaction *transaction.Transaction) error {
	if m.AtomicUpdateBalanceAndCreateTransactionFunc != nil {
		return m.AtomicUpdateBalanceAndCreateTransactionFunc(userID, newBalance, newTransaction)
	}
	return errors.New("AtomicUpdateBalanceAndCreateTransactionFunc not set")
}

func (m *MockUserRepository) Create(user *user.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(user)
	}
	return errors.New("CreateFunc not set")
}
