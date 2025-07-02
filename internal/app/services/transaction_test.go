package services_test

import (
	"database/sql"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/zaynkorai/enlabs/internal/app/services"
	"github.com/zaynkorai/enlabs/internal/domain/transaction"
	"github.com/zaynkorai/enlabs/internal/domain/user"
	"github.com/zaynkorai/enlabs/internal/mocks"
	appErrors "github.com/zaynkorai/enlabs/pkg/errors"
)

func TestTransactionService_ProcessTransaction_Win(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTransactionRepo := &mocks.MockTransactionRepository{}
	svc := services.NewTransactionService(mockUserRepo, mockTransactionRepo)

	userID := uint64(1)
	initialBalance := decimal.NewFromFloat(100.00)
	winAmount := decimal.NewFromFloat(10.50)
	expectedBalance := initialBalance.Add(winAmount)

	mockUserRepo.GetByIDFunc = func(id uint64) (*user.User, error) {
		if id == userID {
			return &user.User{ID: userID, Balance: initialBalance}, nil
		}
		return nil, sql.ErrNoRows
	}

	mockTransactionRepo.GetByTransactionIDFunc = func(transactionID string) (*transaction.Transaction, error) {
		return nil, sql.ErrNoRows // No existing transaction
	}

	mockUserRepo.AtomicUpdateBalanceAndCreateTransactionFunc = func(
		uid uint64,
		newBalance decimal.Decimal,
		newTxn *transaction.Transaction,
	) error {
		assert.Equal(t, userID, uid)
		assert.True(t, newBalance.Equal(expectedBalance))
		assert.Equal(t, "win", newTxn.State)
		assert.True(t, newTxn.Amount.Equal(winAmount))
		return nil
	}

	reqTransaction := &transaction.Transaction{
		TransactionID: "txn-win-1",
		State:         "win",
		Amount:        winAmount,
	}

	err := svc.ProcessTransaction(userID, reqTransaction)
	assert.NoError(t, err)
}

func TestTransactionService_ProcessTransaction_Lose_SufficientBalance(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTransactionRepo := &mocks.MockTransactionRepository{}
	svc := services.NewTransactionService(mockUserRepo, mockTransactionRepo)

	userID := uint64(1)
	initialBalance := decimal.NewFromFloat(100.00)
	loseAmount := decimal.NewFromFloat(10.50)
	expectedBalance := initialBalance.Sub(loseAmount)

	mockUserRepo.GetByIDFunc = func(id uint64) (*user.User, error) {
		if id == userID {
			return &user.User{ID: userID, Balance: initialBalance}, nil
		}
		return nil, sql.ErrNoRows
	}

	mockTransactionRepo.GetByTransactionIDFunc = func(transactionID string) (*transaction.Transaction, error) {
		return nil, sql.ErrNoRows // No existing transaction
	}

	mockUserRepo.AtomicUpdateBalanceAndCreateTransactionFunc = func(
		uid uint64,
		newBalance decimal.Decimal,
		newTxn *transaction.Transaction,
	) error {
		assert.Equal(t, userID, uid)
		assert.True(t, newBalance.Equal(expectedBalance))
		assert.Equal(t, "lose", newTxn.State)
		assert.True(t, newTxn.Amount.Equal(loseAmount))
		return nil
	}

	reqTransaction := &transaction.Transaction{
		TransactionID: "txn-lose-1",
		State:         "lose",
		Amount:        loseAmount,
		SourceType:    "game",
	}

	err := svc.ProcessTransaction(userID, reqTransaction)
	assert.NoError(t, err)
}

func TestTransactionService_ProcessTransaction_Lose_InsufficientBalance(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTransactionRepo := &mocks.MockTransactionRepository{}
	svc := services.NewTransactionService(mockUserRepo, mockTransactionRepo)

	userID := uint64(1)
	initialBalance := decimal.NewFromFloat(5.00)
	loseAmount := decimal.NewFromFloat(10.50) // More than initial balance

	mockUserRepo.GetByIDFunc = func(id uint64) (*user.User, error) {
		if id == userID {
			return &user.User{ID: userID, Balance: initialBalance}, nil
		}
		return nil, sql.ErrNoRows
	}

	mockTransactionRepo.GetByTransactionIDFunc = func(transactionID string) (*transaction.Transaction, error) {
		return nil, sql.ErrNoRows // No existing transaction
	}

	// AtomicUpdateBalanceAndCreateTransactionFunc should not be called in this case
	mockUserRepo.AtomicUpdateBalanceAndCreateTransactionFunc = func(
		uid uint64,
		newBalance decimal.Decimal,
		newTxn *transaction.Transaction,
	) error {
		t.Fatal("AtomicUpdateBalanceAndCreateTransactionFunc should not be called for insufficient balance")
		return nil
	}

	reqTransaction := &transaction.Transaction{
		TransactionID: "txn-lose-insufficient",
		State:         "lose",
		Amount:        loseAmount,
		SourceType:    "game",
	}

	err := svc.ProcessTransaction(userID, reqTransaction)
	assert.Error(t, err)
	assert.True(t, appErrors.IsValidationError(err))
	assert.Contains(t, err.Error(), "insufficient balance")
}

func TestTransactionService_ProcessTransaction_DuplicateTransactionID(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTransactionRepo := &mocks.MockTransactionRepository{}
	svc := services.NewTransactionService(mockUserRepo, mockTransactionRepo)

	userID := uint64(1)
	existingTransactionID := "duplicate-txn-id"

	mockTransactionRepo.GetByTransactionIDFunc = func(transactionID string) (*transaction.Transaction, error) {
		if transactionID == existingTransactionID {
			return &transaction.Transaction{TransactionID: existingTransactionID}, nil // Transaction already exists
		}
		return nil, sql.ErrNoRows
	}

	// GetByIDFunc and AtomicUpdateBalanceAndCreateTransactionFunc should not be called
	mockUserRepo.GetByIDFunc = func(id uint64) (*user.User, error) {
		t.Fatal("GetByIDFunc should not be called for duplicate transaction")
		return nil, nil
	}
	mockUserRepo.AtomicUpdateBalanceAndCreateTransactionFunc = func(
		uid uint64,
		newBalance decimal.Decimal,
		newTxn *transaction.Transaction,
	) error {
		t.Fatal("AtomicUpdateBalanceAndCreateTransactionFunc should not be called for duplicate transaction")
		return nil
	}

	reqTransaction := &transaction.Transaction{
		TransactionID: existingTransactionID,
		State:         "win",
		Amount:        decimal.NewFromFloat(10.00),
		SourceType:    "game",
	}

	err := svc.ProcessTransaction(userID, reqTransaction)
	assert.Error(t, err)
	assert.True(t, appErrors.IsConflictError(err))
	assert.Contains(t, err.Error(), "transaction with this ID has already been processed")
}

func TestTransactionService_ProcessTransaction_UserNotFound(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTransactionRepo := &mocks.MockTransactionRepository{}
	svc := services.NewTransactionService(mockUserRepo, mockTransactionRepo)

	userID := uint64(999) // Non-existent user

	mockTransactionRepo.GetByTransactionIDFunc = func(transactionID string) (*transaction.Transaction, error) {
		return nil, sql.ErrNoRows // No existing transaction
	}

	mockUserRepo.GetByIDFunc = func(id uint64) (*user.User, error) {
		return nil, sql.ErrNoRows // User not found
	}

	// AtomicUpdateBalanceAndCreateTransactionFunc should not be called
	mockUserRepo.AtomicUpdateBalanceAndCreateTransactionFunc = func(
		uid uint64,
		newBalance decimal.Decimal,
		newTxn *transaction.Transaction,
	) error {
		t.Fatal("AtomicUpdateBalanceAndCreateTransactionFunc should not be called for user not found")
		return nil
	}

	reqTransaction := &transaction.Transaction{
		TransactionID: "txn-user-not-found",
		State:         "win",
		Amount:        decimal.NewFromFloat(5.00),
		SourceType:    "game",
	}

	err := svc.ProcessTransaction(userID, reqTransaction)
	assert.Error(t, err)
	assert.True(t, appErrors.IsNotFoundError(err))
	assert.Contains(t, err.Error(), "user with ID 999 not found")
}

func TestTransactionService_GetUserBalance_Success(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTransactionRepo := &mocks.MockTransactionRepository{}
	svc := services.NewTransactionService(mockUserRepo, mockTransactionRepo)

	userID := uint64(1)
	expectedBalance := decimal.NewFromFloat(123.45)

	mockUserRepo.GetByIDFunc = func(id uint64) (*user.User, error) {
		if id == userID {
			return &user.User{ID: userID, Balance: expectedBalance}, nil
		}
		return nil, sql.ErrNoRows
	}

	u, err := svc.GetUserBalance(userID)
	assert.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, userID, u.ID)
	assert.True(t, u.Balance.Equal(expectedBalance))
}
func TestTransactionService_ProcessTransaction_InvalidState(t *testing.T) {
	mockUserRepo := &mocks.MockUserRepository{}
	mockTransactionRepo := &mocks.MockTransactionRepository{}
	svc := services.NewTransactionService(mockUserRepo, mockTransactionRepo)

	userID := uint64(1)
	initialBalance := decimal.NewFromFloat(100.00)

	mockUserRepo.GetByIDFunc = func(id uint64) (*user.User, error) {
		if id == userID {
			return &user.User{ID: userID, Balance: initialBalance}, nil
		}
		return nil, sql.ErrNoRows
	}

	mockTransactionRepo.GetByTransactionIDFunc = func(transactionID string) (*transaction.Transaction, error) {
		return nil, sql.ErrNoRows
	}

	reqTransaction := &transaction.Transaction{
		TransactionID: "txn-invalid-state",
		State:         "invalid", // Invalid state
		Amount:        decimal.NewFromFloat(10.00),
		SourceType:    "game",
	}

	err := svc.ProcessTransaction(userID, reqTransaction)
	assert.Error(t, err)
	assert.True(t, appErrors.IsValidationError(err))
	assert.Contains(t, err.Error(), "invalid transaction state")
}
