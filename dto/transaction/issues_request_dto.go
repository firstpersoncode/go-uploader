package dto_transaction

type PaginationDTO struct {
	Page  int `query:"page"`
	Limit int `query:"limit"`
}

type SortDirection string

const (
	SortAsc  SortDirection = "ASC"
	SortDesc SortDirection = "DESC"
)

type SortingDTO struct {
	Sort   SortDirection `query:"sort"`
	SortBy string        `query:"sortBy"`
}
