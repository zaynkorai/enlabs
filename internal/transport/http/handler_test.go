package http_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/zaynkorai/enlabs/internal/app/services"
	"github.com/zaynkorai/enlabs/internal/domain/user"
	"github.com/zaynkorai/enlabs/internal/platform/persistence"
	apihandler "github.com/zaynkorai/enlabs/internal/transport/http"
	"github.com/zaynkorai/enlabs/pkg/config"
	"github.com/zaynkorai/enlabs/pkg/database"
	"gorm.io/gorm"
)

var (
	testDB    *gorm.DB
	router    *gin.Engine
	userRepo  *persistence.UserRepository
	txnRepo   *persistence.TransactionRepository
	testUsers = []uint64{1, 2, 3} // Predefined users
)

func TestMain(m *testing.M) {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load test configuration: %v\n", err)
		os.Exit(1)
	}

	testDB, err = database.NewPostgresDB(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to test database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Test database connected and migrations run.")

	userRepo = persistence.NewUserRepository(testDB)
	txnRepo = persistence.NewTransactionRepository(testDB)
	transactionService := services.NewTransactionService(userRepo, txnRepo)

	gin.SetMode(gin.TestMode)
	router = gin.Default()
	handler := apihandler.NewHandler(transactionService)
	router.POST("/user/:userId/transaction", handler.ProcessTransaction)
	router.GET("/user/:userId/balance", handler.GetUserBalance)

	exitCode := m.Run()

	sqlDB, _ := testDB.DB()
	if sqlDB != nil {
		sqlDB.Close()
		fmt.Println("Test database connection closed.")
	}
	os.Exit(exitCode)
}

func setupTest(t *testing.T) {
	err := testDB.Exec("TRUNCATE TABLE transactions RESTART IDENTITY CASCADE").Error
	assert.NoError(t, err, "Failed to truncate transactions table")
	err = testDB.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE").Error
	assert.NoError(t, err, "Failed to truncate users table")

	for _, id := range testUsers {
		initialUser := user.User{ID: id, Balance: decimal.NewFromInt(0)}
		err := userRepo.Create(&initialUser)
		assert.NoError(t, err, "Failed to re-seed user %d", id)
	}
}

func TestProcessTransaction_Win_Success(t *testing.T) {
	setupTest(t)
	userID := testUsers[0]
	amount := "10.50"
	transactionID := "test-txn-win-1"

	body := apihandler.TransactionRequest{
		State:         "win",
		Amount:        amount,
		TransactionID: transactionID,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/user/%d/transaction", userID),
		bytes.NewBuffer(jsonBody),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Source-Type", "game")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "{}", w.Body.String())

	userBalance, err := userRepo.GetByID(userID)
	assert.NoError(t, err)
	assert.True(t, userBalance.Balance.Equal(decimal.NewFromFloat(10.50)))
}

func TestProcessTransaction_Lose_Success(t *testing.T) {
	setupTest(t)
	userID := testUsers[1]
	initialBalance := decimal.NewFromFloat(50.00)
	err := testDB.Model(&user.User{}).Where("id = ?", userID).Update("balance", initialBalance).Error
	assert.NoError(t, err)

	amount := "25.75"
	transactionID := "test-txn-lose-1"

	body := apihandler.TransactionRequest{
		State:         "lose",
		Amount:        amount,
		TransactionID: transactionID,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/user/%d/transaction", userID),
		bytes.NewBuffer(jsonBody),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Source-Type", "payment")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	userBalance, err := userRepo.GetByID(userID)
	assert.NoError(t, err)
	expectedBalance := initialBalance.Sub(decimal.NewFromFloat(25.75))
	assert.True(t, userBalance.Balance.Equal(expectedBalance))
}

func TestProcessTransaction_Lose_InsufficientBalance(t *testing.T) {
	setupTest(t)
	userID := testUsers[2]
	amount := "10.00"
	transactionID := "test-txn-lose-insufficient"

	body := apihandler.TransactionRequest{
		State:         "lose",
		Amount:        amount,
		TransactionID: transactionID,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/user/%d/transaction", userID),
		bytes.NewBuffer(jsonBody),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Source-Type", "game")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var responseBody map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Contains(t, responseBody["error"], "insufficient balance")
	assert.Contains(t, responseBody["error"], "remains 0.00")
	userBalance, err := userRepo.GetByID(userID)
	assert.NoError(t, err)
	assert.True(t, userBalance.Balance.Equal(decimal.Zero))
}

func TestProcessTransaction_DuplicateTransactionID(t *testing.T) {
	setupTest(t)
	userID := testUsers[0]
	amount := "5.00"
	transactionID := "duplicate-txn-id"

	body1 := apihandler.TransactionRequest{State: "win", Amount: amount, TransactionID: transactionID}
	jsonBody1, _ := json.Marshal(body1)
	req1 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/user/%d/transaction", userID), bytes.NewBuffer(jsonBody1))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Source-Type", "game")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	body2 := apihandler.TransactionRequest{State: "win", Amount: "1.00", TransactionID: transactionID} // Different amount to ensure idempotency check works
	jsonBody2, _ := json.Marshal(body2)
	req2 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/user/%d/transaction", userID), bytes.NewBuffer(jsonBody2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Source-Type", "game")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusConflict, w2.Code)
	var responseBody map[string]string
	err := json.Unmarshal(w2.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Contains(t, responseBody["error"], "transaction with this ID has already been processed")
	userBalance, err := userRepo.GetByID(userID)
	assert.NoError(t, err)
	assert.True(t, userBalance.Balance.Equal(decimal.NewFromFloat(5.00)))
}

func TestGetUserBalance_Success(t *testing.T) {
	setupTest(t)
	userID := testUsers[0]
	initialBalance := decimal.NewFromFloat(99.99)
	err := testDB.Model(&user.User{}).Where("id = ?", userID).Update("balance", initialBalance).Error
	assert.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/user/%d/balance", userID),
		nil,
	)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var responseBody apihandler.BalanceResponse
	err = json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Equal(t, userID, responseBody.UserID)
	assert.Equal(t, "99.99", responseBody.Balance)
}

func TestProcessTransaction_InvalidUserID(t *testing.T) {
	setupTest(t)
	invalidUserID := "abc"

	body := apihandler.TransactionRequest{State: "win", Amount: "10.00", TransactionID: "txn-invalid-user-id"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/user/%s/transaction", invalidUserID),
		bytes.NewBuffer(jsonBody),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Source-Type", "game")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var responseBody map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Contains(t, responseBody["error"], "Invalid userId. Must be a positive integer.")
}

func TestProcessTransaction_MissingSourceTypeHeader(t *testing.T) {
	setupTest(t)
	userID := testUsers[0]

	body := apihandler.TransactionRequest{State: "win", Amount: "10.00", TransactionID: "txn-missing-source-type"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/user/%d/transaction", userID),
		bytes.NewBuffer(jsonBody),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var responseBody map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Contains(t, responseBody["error"], "Missing Source-Type header")
}

func TestProcessTransaction_InvalidStateInBody(t *testing.T) {
	setupTest(t)
	userID := testUsers[0]

	body := apihandler.TransactionRequest{State: "unknown", Amount: "10.00", TransactionID: "txn-invalid-state"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/user/%d/transaction", userID),
		bytes.NewBuffer(jsonBody),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Source-Type", "game")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var responseBody map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Contains(t, responseBody["error"], "Error:Field validation for 'State' failed on the 'oneof' tag")
}

func TestProcessTransaction_InvalidAmountFormat(t *testing.T) {
	setupTest(t)
	userID := testUsers[0]

	body := apihandler.TransactionRequest{State: "win", Amount: "invalid-amount", TransactionID: "txn-invalid-amount"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/user/%d/transaction", userID),
		bytes.NewBuffer(jsonBody),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Source-Type", "game")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var responseBody map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Contains(t, responseBody["error"], "Invalid amount format. Must be a valid decimal string.")
}

func TestProcessTransaction_NegativeOrZeroAmount(t *testing.T) {
	setupTest(t)
	userID := testUsers[0]

	testCases := []struct {
		amount string
		txnID  string
	}{
		{"-5.00", "txn-negative-amount"},
		{"0.00", "txn-zero-amount"},
	}

	for _, tc := range testCases {
		t.Run("Amount_"+tc.amount, func(t *testing.T) {
			body := apihandler.TransactionRequest{State: "win", Amount: tc.amount, TransactionID: tc.txnID}
			jsonBody, _ := json.Marshal(body)

			req := httptest.NewRequest(
				http.MethodPost,
				fmt.Sprintf("/user/%d/transaction", userID),
				bytes.NewBuffer(jsonBody),
			)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Source-Type", "game")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			var responseBody map[string]string
			err := json.Unmarshal(w.Body.Bytes(), &responseBody)
			assert.NoError(t, err)
			assert.Contains(t, responseBody["error"], "Amount must be positive.")
		})
	}
}
