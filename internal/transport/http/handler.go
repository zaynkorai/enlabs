package http

import (
	"log"
	"net/http"
	"strconv"

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
	return &Handler{
		transactionService: transactionService,
		validator:          validator.New(),
	}
}

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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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

	newTransaction := &transaction.Transaction{
		UserID:        userID,
		TransactionID: req.TransactionID,
		SourceType:    sourceType,
		State:         req.State,
		Amount:        amount,
	}

	err = h.transactionService.ProcessTransaction(userID, newTransaction)
	if err != nil {
		log.Printf("Error processing transaction for user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.Status(http.StatusOK)
}

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
