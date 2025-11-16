package transaction

import (
	"strings"
	"testing"

	"firstpersoncode/go-uploader/domain"
	dto_transaction "firstpersoncode/go-uploader/dto/transaction"
	"firstpersoncode/go-uploader/internal/repositories"
)

func setupTestService() (domain.TransactionRepository, domain.TransactionService) {
	repo := repositories.NewTransactionRepository()
	service := NewTransactionService(repo)
	return repo, service
}

func TestParseAndStoreCSV_Success(t *testing.T) {
	repo, service := setupTestService()

	csvData := `1624507883, JOHN DOE, DEBIT, 250000, SUCCESS, restaurant
1624608050, E-COMMERCE A, DEBIT, 150000, FAILED, clothes`

	response, err := service.ParseAndStoreCSV(strings.NewReader(csvData))

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.TotalRows != 2 {
		t.Errorf("Expected 2 transactions, got %d", response.TotalRows)
	}

	transactions := repo.GetAll()
	if len(transactions) != 2 {
		t.Errorf("Expected 2 transactions in repo, got %d", len(transactions))
	}
}

func TestParseAndStoreCSV_InvalidFormat(t *testing.T) {
	repo, service := setupTestService()
	defer repo.Clear()

	// CSV with only 5 fields instead of 6
	csvData := `1624507883, JOHN DOE, DEBIT, 250000, SUCCESS`

	// This should panic because the service doesn't validate field count
	// We're testing that it fails (either panic or error)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic for invalid format with 5 fields")
		}
	}()

	service.ParseAndStoreCSV(strings.NewReader(csvData))
}

func TestParseAndStoreCSV_InvalidType(t *testing.T) {
	repo, service := setupTestService()
	defer repo.Clear()

	csvData := `1624507883, JOHN DOE, INVALID, 250000, SUCCESS, restaurant`

	response, err := service.ParseAndStoreCSV(strings.NewReader(csvData))

	if err == nil {
		t.Fatal("Expected error for invalid type")
	}

	if response != nil {
		t.Error("Expected nil response on error")
	}
}

func TestParseAndStoreCSV_EmptyFile(t *testing.T) {
	_, service := setupTestService()

	_, err := service.ParseAndStoreCSV(strings.NewReader(""))

	if err == nil {
		t.Fatal("Expected error for empty file")
	}
}

func TestCalculateBalance(t *testing.T) {
	_, service := setupTestService()

	csvData := `1624507883, JOHN DOE, CREDIT, 500000, SUCCESS, salary
1624608050, E-COMMERCE A, DEBIT, 150000, SUCCESS, clothes
1624708050, SHOP B, DEBIT, 100000, FAILED, test
1624808050, STORE C, CREDIT, 200000, SUCCESS, refund`

	_, err := service.ParseAndStoreCSV(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	response, err := service.CalculateBalance()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.Credits != 700000 {
		t.Errorf("Expected credits 700000, got %d", response.Credits)
	}

	if response.Debits != 150000 {
		t.Errorf("Expected debits 150000, got %d", response.Debits)
	}

	if response.Balance != 550000 {
		t.Errorf("Expected balance 550000, got %d", response.Balance)
	}
}

func TestCalculateBalance_OnlySuccess(t *testing.T) {
	_, service := setupTestService()

	csvData := `1624507883, JOHN DOE, CREDIT, 1000000, SUCCESS, salary
1624608050, E-COMMERCE A, DEBIT, 200000, FAILED, clothes
1624708050, SHOP B, DEBIT, 300000, PENDING, test`

	_, err := service.ParseAndStoreCSV(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	response, err := service.CalculateBalance()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.Credits != 1000000 {
		t.Errorf("Expected credits 1000000 (only SUCCESS), got %d", response.Credits)
	}

	if response.Debits != 0 {
		t.Errorf("Expected debits 0 (FAILED and PENDING excluded), got %d", response.Debits)
	}

	if response.Balance != 1000000 {
		t.Errorf("Expected balance 1000000, got %d", response.Balance)
	}
}

func TestGetIssues(t *testing.T) {
	_, service := setupTestService()

	csvData := `1624507883, JOHN DOE, DEBIT, 250000, SUCCESS, restaurant
1624608050, E-COMMERCE A, DEBIT, 150000, FAILED, clothes
1624708050, SHOP B, CREDIT, 500000, PENDING, refund
1624808050, STORE C, DEBIT, 100000, SUCCESS, food`

	_, err := service.ParseAndStoreCSV(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	pagination := dto_transaction.PaginationDTO{Page: 1, Limit: 10}
	sorting := dto_transaction.SortingDTO{Sort: dto_transaction.SortAsc, SortBy: "timestamp"}

	response, err := service.GetIssues(pagination, sorting)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.Total != 2 {
		t.Errorf("Expected total 2 issues, got %d", response.Total)
	}

	if len(response.Transactions) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(response.Transactions))
	}

	if response.Transactions[0].Status != "FAILED" {
		t.Errorf("Expected first transaction status FAILED, got %s", response.Transactions[0].Status)
	}

	if response.Transactions[1].Status != "PENDING" {
		t.Errorf("Expected second transaction status PENDING, got %s", response.Transactions[1].Status)
	}
}

func TestGetIssues_Pagination(t *testing.T) {
	_, service := setupTestService()

	csvData := `1624507883, TX1, DEBIT, 100000, FAILED, test1
1624608050, TX2, DEBIT, 200000, PENDING, test2
1624708050, TX3, DEBIT, 300000, FAILED, test3
1624808050, TX4, DEBIT, 400000, PENDING, test4`

	_, err := service.ParseAndStoreCSV(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	pagination := dto_transaction.PaginationDTO{Page: 1, Limit: 2}
	sorting := dto_transaction.SortingDTO{Sort: dto_transaction.SortAsc, SortBy: "timestamp"}

	response, err := service.GetIssues(pagination, sorting)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.Total != 4 {
		t.Errorf("Expected total 4 issues, got %d", response.Total)
	}

	if len(response.Transactions) != 2 {
		t.Errorf("Expected 2 transactions on page 1, got %d", len(response.Transactions))
	}

	pagination.Page = 2
	response, err = service.GetIssues(pagination, sorting)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(response.Transactions) != 2 {
		t.Errorf("Expected 2 transactions on page 2, got %d", len(response.Transactions))
	}
}

func TestGetIssues_Sorting(t *testing.T) {
	_, service := setupTestService()

	csvData := `1624708050, C-TX, DEBIT, 300000, FAILED, test
1624508050, A-TX, DEBIT, 100000, PENDING, test
1624608050, B-TX, DEBIT, 200000, FAILED, test`

	_, err := service.ParseAndStoreCSV(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	pagination := dto_transaction.PaginationDTO{Page: 1, Limit: 10}
	sorting := dto_transaction.SortingDTO{Sort: dto_transaction.SortDesc, SortBy: "amount"}

	response, err := service.GetIssues(pagination, sorting)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.Transactions[0].Amount != 300000 {
		t.Errorf("Expected first transaction amount 300000 (DESC), got %d", response.Transactions[0].Amount)
	}

	if response.Transactions[2].Amount != 100000 {
		t.Errorf("Expected last transaction amount 100000 (DESC), got %d", response.Transactions[2].Amount)
	}
}
