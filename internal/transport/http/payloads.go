package http

// TransactionRequest represents the incoming JSON payload for a transaction.
// @Description Details for a new transaction to update user balance.
type TransactionRequest struct {
	State         string `json:"state" binding:"required,oneof=win lose"`
	Amount        string `json:"amount" binding:"required"`
	TransactionID string `json:"transactionId" binding:"required"`
}

// BalanceResponse represents the JSON payload for getting user balance.
// @Description Current user balance information.
type BalanceResponse struct {
	UserID  uint64 `json:"userId"`
	Balance string `json:"balance"`
}
