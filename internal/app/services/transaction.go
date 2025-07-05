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
		// Pre-check if balance would go negative.
		// This client-side check prevents unnecessary database transactions for invalid requests.
		if user.Balance.LessThan(reqTransaction.Amount) {
			return appErrors.NewValidationError("insufficient balance")
		}
		newBalance = user.Balance.Sub(reqTransaction.Amount)
	default:
		return appErrors.NewValidationError("invalid transaction state")
	}

	err = s.userRepo.AtomicUpdateBalanceAndCreateTransaction(user.ID, newBalance, reqTransaction)
	if err != nil {
		if appErrors.IsAlreadyProcessedError(err) {

			log.Printf("Transaction ID %s for user %d already processed. Skipping balance update.", reqTransaction.TransactionID, userID)
			return nil // Return nil to indicate success to the caller (HTTP handler)
		}
		if appErrors.IsNotFoundError(err) {

			return err
		}

		return fmt.Errorf("failed to update user balance and record transaction atomically: %w", err)
	}

	log.Printf("User %d balance updated to %s for transaction %s", userID, newBalance.StringFixed(2), reqTransaction.TransactionID)
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
