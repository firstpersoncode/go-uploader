package repositories

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"firstpersoncode/go-uploader/domain"
	dto_transaction "firstpersoncode/go-uploader/dto/transaction"
)

type transactionRepository struct {
	mu           sync.RWMutex
	transactions []domain.Transaction
}

func NewTransactionRepository() domain.TransactionRepository {
	return &transactionRepository{
		transactions: make([]domain.Transaction, 0),
	}
}

func (r *transactionRepository) SaveAll(transactions []domain.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for index, tx := range transactions {
		record := []string{
			strconv.FormatInt(tx.Timestamp.Unix(), 10),
			tx.Name,
			string(tx.Type),
			strconv.FormatInt(tx.Amount, 10),
			string(tx.Status),
			tx.Description,
		}

		if err := r.validateRecord(record); err != nil {
			return fmt.Errorf("line %d: %v", index, err)
		}
	}

	r.transactions = append(r.transactions, transactions...)
	return nil
}

func (r *transactionRepository) GetAll() []domain.Transaction {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.transactions
}

func (r *transactionRepository) GetAllByUserID(userID string) []domain.Transaction {
	allTransactions := r.GetAll()
	var userTransactions []domain.Transaction
	for _, tx := range allTransactions {
		if tx.UserID == userID {
			userTransactions = append(userTransactions, tx)
		}
	}
	return userTransactions
}

func (r *transactionRepository) GetAllIssues(userID string, pagination dto_transaction.PaginationDTO, sorting dto_transaction.SortingDTO) ([]domain.Transaction, int, error) {
	userTransactions := r.GetAllByUserID(userID)
	var issues []domain.Transaction

	for _, tx := range userTransactions {
		if tx.Status == domain.TransactionStatusFailed || tx.Status == domain.TransactionStatusPending {
			issues = append(issues, tx)
		}
	}

	if sorting.SortBy == "" {
		sorting.SortBy = "timestamp"
	}

	if sorting.Sort == "" {
		sorting.Sort = "ASC"
	}

	err := r.sortTransactions(issues, sorting.Sort, sorting.SortBy)
	if err != nil {
		return nil, 0, err
	}

	if pagination.Page < 1 {
		pagination.Page = 1
	}

	if pagination.Limit < 1 {
		pagination.Limit = 10
	}

	total := len(issues)
	start := (pagination.Page - 1) * pagination.Limit
	end := start + pagination.Limit
	if start >= total {
		return make([]domain.Transaction, 0), 0, nil
	}

	if end > total {
		end = total
	}

	return issues[start:end], total, nil
}

func (r *transactionRepository) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.transactions = make([]domain.Transaction, 0)
}

func (r *transactionRepository) sortTransactions(transactions []domain.Transaction, sort dto_transaction.SortDirection, sortBy string) error {
	if len(transactions) == 0 {
		return nil
	}

	ascending := strings.ToUpper(string(sort)) != "DESC"

	for i := 0; i < len(transactions)-1; i++ {
		for j := i + 1; j < len(transactions); j++ {
			shouldSwap := false

			switch strings.ToLower(sortBy) {
			case "timestamp":
				if ascending {
					shouldSwap = transactions[i].Timestamp.After(transactions[j].Timestamp)
				} else {
					shouldSwap = transactions[i].Timestamp.Before(transactions[j].Timestamp)
				}
			case "name":
				if ascending {
					shouldSwap = transactions[i].Name > transactions[j].Name
				} else {
					shouldSwap = transactions[i].Name < transactions[j].Name
				}
			case "amount":
				if ascending {
					shouldSwap = transactions[i].Amount > transactions[j].Amount
				} else {
					shouldSwap = transactions[i].Amount < transactions[j].Amount
				}
			case "type":
				if ascending {
					shouldSwap = transactions[i].Type > transactions[j].Type
				} else {
					shouldSwap = transactions[i].Type < transactions[j].Type
				}
			case "status":
				if ascending {
					shouldSwap = transactions[i].Status > transactions[j].Status
				} else {
					shouldSwap = transactions[i].Status < transactions[j].Status
				}
			default:
				return fmt.Errorf("invalid sortBy field: %s", sortBy)
			}

			if shouldSwap {
				transactions[i], transactions[j] = transactions[j], transactions[i]
			}
		}
	}

	return nil
}

func (r *transactionRepository) validateRecord(record []string) error {
	if len(record) != 6 {
		return fmt.Errorf("expected 6 fields, got %d", len(record))
	}

	if strings.TrimSpace(record[0]) == "" {
		return fmt.Errorf("timestamp is required")
	}
	if strings.TrimSpace(record[1]) == "" {
		return fmt.Errorf("name is required")
	}
	if strings.TrimSpace(record[2]) == "" {
		return fmt.Errorf("type is required")
	}
	if strings.TrimSpace(record[3]) == "" {
		return fmt.Errorf("amount is required")
	}
	if strings.TrimSpace(record[4]) == "" {
		return fmt.Errorf("status is required")
	}
	if strings.TrimSpace(record[5]) == "" {
		return fmt.Errorf("description is required")
	}

	// Validate type
	txType := strings.ToUpper(strings.TrimSpace(record[2]))
	if txType != string(domain.TransactionTypeDebit) && txType != string(domain.TransactionTypeCredit) {
		return fmt.Errorf("invalid type")
	}

	status := strings.ToUpper(strings.TrimSpace(record[4]))
	if status != string(domain.TransactionStatusSuccess) && status != string(domain.TransactionStatusFailed) && status != string(domain.TransactionStatusPending) {
		return fmt.Errorf("invalid status")
	}

	timestamp, err := strconv.ParseInt(strings.TrimSpace(record[0]), 10, 64)
	if err != nil || timestamp <= 0 {
		return fmt.Errorf("invalid timestamp")
	}

	amount, err := strconv.ParseInt(strings.TrimSpace(record[3]), 10, 64)
	if err != nil || amount < 0 {
		return fmt.Errorf("invalid amount")
	}

	return nil
}
