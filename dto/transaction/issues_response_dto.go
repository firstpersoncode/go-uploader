package dto_transaction

type IssuesResponseDTO struct {
	Transactions []TransactionDTO `json:"transactions"`
	Total        int              `json:"total"`
}

type TransactionDTO struct {
	Timestamp   string `json:"timestamp"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Amount      int64  `json:"amount"`
	Status      string `json:"status"`
	Description string `json:"description"`
}
