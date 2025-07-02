package persistence

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
	"github.com/zaynkorai/enlabs/internal/domain/transaction"
	"github.com/zaynkorai/enlabs/internal/domain/user"
	appErrors "github.com/zaynkorai/enlabs/pkg/errors"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(id uint64) (*user.User, error) {
	var u user.User
	result := r.db.First(&u, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, sql.ErrNoRows // Conform to standard library error for "not found"
		}
		return nil, fmt.Errorf("failed to get user by ID %d: %w", id, result.Error)
	}
	return &u, nil
}

// AtomicUpdateBalanceAndCreateTransaction performs both operations in a single database transaction
// to ensure atomicity and consistency.
func (r *UserRepository) AtomicUpdateBalanceAndCreateTransaction(userID uint64, newBalance decimal.Decimal, newTransaction *transaction.Transaction) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&user.User{}).
			Where("id = ?", userID).
			Updates(map[string]interface{}{"balance": newBalance, "updated_at": gorm.Expr("NOW()")})

		if result.Error != nil {
			return fmt.Errorf("failed to update user balance: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return appErrors.NewNotFoundError(fmt.Sprintf("user with ID %d not found during balance update", userID))
		}

		newTransaction.UserID = userID
		if createErr := tx.Create(newTransaction).Error; createErr != nil {
			var pgErr *pgconn.PgError
			if errors.As(createErr, &pgErr) && pgErr.Code == "23505" { // 23505 is unique_violation
				return appErrors.NewConflictError("transaction with this ID has already been processed (duplicate key)")
			}
			return fmt.Errorf("failed to create transaction record: %w", createErr)
		}

		return nil
	})
}

func (r *UserRepository) Create(user *user.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}
