package domain

import (
	dto_transaction "firstpersoncode/go-uploader/dto/transaction"
	"io"
	"time"

	"github.com/gofiber/fiber/v2"
)

type TransactionType string

const (
	TransactionTypeDebit  TransactionType = "DEBIT"
	TransactionTypeCredit TransactionType = "CREDIT"
)

type TransactionStatus string

const (
	TransactionStatusSuccess TransactionStatus = "SUCCESS"
	TransactionStatusFailed  TransactionStatus = "FAILED"
	TransactionStatusPending TransactionStatus = "PENDING"
)

type Transaction struct {
	Timestamp   time.Time         `json:"timestamp"`
	Name        string            `json:"name"`
	Type        TransactionType   `json:"type"`
	Amount      int64             `json:"amount"`
	Status      TransactionStatus `json:"status"`
	Description string            `json:"description"`
	UserID      string            `json:"user_id"`
}

type TransactionRepository interface {
	SaveAll(transactions []Transaction) error
	GetAll() []Transaction
	GetAllByUserID(userID string) []Transaction
	GetAllIssues(userID string, pagination dto_transaction.PaginationDTO, sorting dto_transaction.SortingDTO) ([]Transaction, int, error)
	Clear()
}

type TransactionService interface {
	ParseAndStoreCSV(fileContent io.Reader, userID string) (*dto_transaction.UploadResponseDTO, error)
	CalculateBalance(userID string) (*dto_transaction.BalanceResponseDTO, error)
	GetIssues(pagination dto_transaction.PaginationDTO, sorting dto_transaction.SortingDTO, userID string) (*dto_transaction.IssuesResponseDTO, error)
}

type TransactionHandler interface {
	UploadStatement(ctx *fiber.Ctx) error
	GetBalance(ctx *fiber.Ctx) error
	GetIssues(ctx *fiber.Ctx) error
}
