package transaction

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"firstpersoncode/go-uploader/domain"
	dto_transaction "firstpersoncode/go-uploader/dto/transaction"
)

type transactionService struct {
	repo domain.TransactionRepository
}

func NewTransactionService(repo domain.TransactionRepository) domain.TransactionService {
	return &transactionService{repo: repo}
}

func (s *transactionService) ParseAndStoreCSV(fileContent io.Reader) (*dto_transaction.UploadResponseDTO, error) {
	reader := csv.NewReader(fileContent)
	reader.TrimLeadingSpace = true

	var transactions []domain.Transaction

	for lineNum := 0; ; lineNum++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("line %d: %v", lineNum, err)
		}

		timestamp, _ := strconv.ParseInt(strings.TrimSpace(record[0]), 10, 64)
		amount, _ := strconv.ParseInt(strings.TrimSpace(record[3]), 10, 64)

		transaction := domain.Transaction{
			Timestamp:   time.Unix(timestamp, 0),
			Name:        strings.TrimSpace(record[1]),
			Type:        domain.TransactionType(strings.ToUpper(strings.TrimSpace(record[2]))),
			Amount:      amount,
			Status:      domain.TransactionStatus(strings.ToUpper(strings.TrimSpace(record[4]))),
			Description: strings.TrimSpace(record[5]),
		}

		transactions = append(transactions, transaction)
	}

	if len(transactions) == 0 {
		return nil, fmt.Errorf("no transactions found")
	}

	if err := s.repo.SaveAll(transactions); err != nil {
		return nil, err
	}

	return &dto_transaction.UploadResponseDTO{
		TotalRows:    len(transactions),
		UploadStatus: "success",
	}, nil
}

func (s *transactionService) CalculateBalance() (*dto_transaction.BalanceResponseDTO, error) {
	transactions := s.repo.GetAll()
	var credits int64 = 0
	var debits int64 = 0
	var balance int64 = 0

	for _, tx := range transactions {
		if tx.Status == domain.TransactionStatusSuccess {
			if tx.Type == domain.TransactionTypeCredit {
				credits += tx.Amount
			} else if tx.Type == domain.TransactionTypeDebit {
				debits += tx.Amount
			}
		}
	}

	balance = credits - debits

	return &dto_transaction.BalanceResponseDTO{
		Credits: credits,
		Debits:  debits,
		Balance: balance,
	}, nil
}

func (s *transactionService) GetIssues(pagination dto_transaction.PaginationDTO, sorting dto_transaction.SortingDTO) (*dto_transaction.IssuesResponseDTO, error) {
	issues, total, err := s.repo.GetAllIssues(pagination, sorting)
	if err != nil {
		return nil, err
	}

	var transactions []dto_transaction.TransactionDTO = make([]dto_transaction.TransactionDTO, 0, len(issues))
	for _, tx := range issues {
		transactions = append(transactions, dto_transaction.TransactionDTO{
			Timestamp:   tx.Timestamp.Format(time.RFC3339),
			Name:        tx.Name,
			Type:        string(tx.Type),
			Amount:      tx.Amount,
			Status:      string(tx.Status),
			Description: tx.Description,
		})
	}

	return &dto_transaction.IssuesResponseDTO{
		Transactions: transactions,
		Total:        total,
	}, nil
}
