package services

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/shopspring/decimal"
	"github.com/zaynkorai/enlabs/internal/domain/transaction"
	"github.com/zaynkorai/enlabs/internal/domain/user"
	appErrors "github.com/zaynkorai/enlabs/pkg/errors"
)

type TransactionService struct {
	userRepo        user.Repository
	transactionRepo transaction.Repository
}

func NewTransactionService(userRepo user.Repository, transactionRepo transaction.Repository) *TransactionService {
	return &TransactionService{
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
	}
}

func (s *TransactionService) ProcessTransaction(userID uint64, reqTransaction *transaction.Transaction) error {

	_, err := s.transactionRepo.GetByTransactionID(reqTransaction.TransactionID) // Check for duplicate transactionId (Idempotency)

	if err == nil {
		return appErrors.NewConflictError("transaction with this ID has already been processed")
	}
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check for existing transaction: %w", err)
	}

	user, err := s.userRepo.GetByID(userID)
	if err == sql.ErrNoRows {
		return appErrors.NewNotFoundError(fmt.Sprintf("user with ID %d not found", userID))
	}
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	var newBalance decimal.Decimal
	switch reqTransaction.State {
	case "win":
		newBalance = user.Balance.Add(reqTransaction.Amount)
	case "lose":
		if user.Balance.LessThan(reqTransaction.Amount) {
			return appErrors.NewValidationError("insufficient balance")
		}
		newBalance = user.Balance.Sub(reqTransaction.Amount)
	default:
		return appErrors.NewValidationError("invalid transaction state")
	}

	err = s.userRepo.AtomicUpdateBalanceAndCreateTransaction(user.ID, newBalance, reqTransaction)
	if err != nil {
		log.Printf("Error in AtomicUpdateBalanceAndCreateTransaction: %v", err)
		if appErrors.IsConflictError(err) {
			return err
		}
		if appErrors.IsValidationError(err) {
			return err
		}
		return fmt.Errorf("failed to update user balance and record transaction: %w", err)
	}

	return nil
}

func (s *TransactionService) GetUserBalance(userID uint64) (*user.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err == sql.ErrNoRows {
		return nil, appErrors.NewNotFoundError(fmt.Sprintf("user with ID %d not found", userID))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user balance: %w", err)
	}
	return user, nil
}
