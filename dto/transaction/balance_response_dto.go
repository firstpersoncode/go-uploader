package dto_transaction

type BalanceResponseDTO struct {
	Credits int64 `json:"credits"`
	Debits  int64 `json:"debits"`
	Balance int64 `json:"balance"`
}
