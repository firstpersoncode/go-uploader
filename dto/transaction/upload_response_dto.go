package dto_transaction

type UploadResponseDTO struct {
	TotalRows    int    `json:"total_rows"`
	UploadStatus string `json:"upload_status"`
}
