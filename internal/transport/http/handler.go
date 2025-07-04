package http

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
	"github.com/zaynkorai/enlabs/internal/app/services"
	"github.com/zaynkorai/enlabs/internal/domain/transaction"
	appErrors "github.com/zaynkorai/enlabs/pkg/errors"
	"github.com/zaynkorai/enlabs/pkg/utils"
)

type Handler struct {
	transactionService *services.TransactionService
	validator          *validator.Validate
}

func NewHandler(transactionService *services.TransactionService) *Handler {
	v := validator.New()

	if err := v.RegisterValidation("decimal_2_places", validateDecimalTwoPlaces); err != nil {
		log.Fatalf("Failed to register custom validator: %v", err)
	}

	return &Handler{
		transactionService: transactionService,
		validator:          v,
	}
}

func validateDecimalTwoPlaces(fl validator.FieldLevel) bool {
	amountStr := fl.Field().String()
	parts := strings.Split(amountStr, ".")

	if len(parts) > 2 { // More than one decimal point
		return false
	}

	if len(parts) == 2 { // Has a decimal part
		decimalPart := parts[1]
		if len(decimalPart) > 2 {
			return false // More than 2 decimal places
		}
	}

	_, err := decimal.NewFromString(amountStr)
	return err == nil
}

// ProcessTransaction
// @Summary Updates user balance based on a transaction
// @Description Processes 'win' or 'lose' transactions and updates the user's balance, ensuring idempotency and non-negative balance.
// @Tags Users
// @Accept json
// @Produce json
// @Param userId path int true "User ID"
// @Param Source-Type header string true "Type of the transaction source (game, server, payment)" Enums(game, server, payment)
// @Param transaction body TransactionRequest true "Transaction details"
// @Success 200 {object} map[string]interface{} "Transaction processed successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid input or insufficient balance"
// @Failure 404 {object} map[string]interface{} "Not Found: User does not exist"
// @Failure 409 {object} map[string]interface{} "Conflict: Transaction with this ID already processed"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /user/{userId}/transaction [post]
func (h *Handler) ProcessTransaction(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil || userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userId. Must be a positive integer."})
		return
	}

	sourceType := c.GetHeader("Source-Type")
	if sourceType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing Source-Type header"})
		return
	}
	if sourceType != "game" && sourceType != "server" && sourceType != "payment" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Source-Type header. Must be 'game', 'server', or 'payment'."})
		return
	}

	var req TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.validator.Struct(req); err != nil {

		log.Printf("Validation error: %v", err)

		if _, ok := err.(*validator.InvalidValidationError); ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal validation error"})
			return
		}
		for _, err := range err.(validator.ValidationErrors) {
			if err.Field() == "Amount" && err.Tag() == "decimal_2_places" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Amount must be a valid number string with up to 2 decimal places."})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed for field '" + err.Field() + "': " + err.Tag()})
			return
		}
	}

	amount, err := utils.ParseDecimal(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount format. Must be a valid decimal string."})
		return
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Amount must be positive."})
		return
	}

	// Ensure amount is rounded to 2 decimal places before passing to service
	// This adds an extra layer of safety, although `validateDecimalTwoPlaces` should prevent >2 places.
	amount = amount.RoundBank(2)

	newTransaction := &transaction.Transaction{
		UserID:        userID,
		TransactionID: req.TransactionID,
		SourceType:    sourceType,
		State:         req.State,
		Amount:        amount,
	}

	err = h.transactionService.ProcessTransaction(userID, newTransaction)
	if err != nil {
		if appErrors.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if appErrors.IsConflictError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if appErrors.IsValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Printf("Error processing transaction for user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.Status(http.StatusOK)
}

// GetUserBalance
// @Summary Gets current user balance
// @Description Retrieves the current balance for a specified user.
// @Tags Users
// @Produce json
// @Param userId path int true "User ID"
// @Success 200 {object} BalanceResponse "Current user balance"
// @Failure 400 {object} map[string]interface{} "Bad Request: Invalid userId"
// @Failure 404 {object} map[string]interface{} "Not Found: User does not exist"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /user/{userId}/balance [get]
func (h *Handler) GetUserBalance(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil || userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userId. Must be a positive number."})
		return
	}

	user, err := h.transactionService.GetUserBalance(userID)
	if err != nil {
		if appErrors.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		log.Printf("Error getting balance for user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, BalanceResponse{
		UserID:  user.ID,
		Balance: user.Balance.StringFixed(2), // Round to 2 decimal places
	})
}
