package transaction

import (
	"strings"

	"firstpersoncode/go-uploader/domain"
	"firstpersoncode/go-uploader/dto"
	dto_transaction "firstpersoncode/go-uploader/dto/transaction"

	"github.com/gofiber/fiber/v2"
)

type transactionHandler struct {
	service domain.TransactionService
}

func NewTransactionHandler(service domain.TransactionService) domain.TransactionHandler {
	return &transactionHandler{service: service}
}

func (api *transactionHandler) UploadStatement(ctx *fiber.Ctx) error {
	file, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(400).JSON(dto.CreateErrorResponse("No file uploaded"))
	}

	if !strings.HasSuffix(file.Filename, ".csv") {
		return ctx.Status(400).JSON(dto.CreateErrorResponse("Only CSV files are allowed"))
	}

	fileContent, err := file.Open()
	if err != nil {
		return ctx.Status(500).JSON(dto.CreateErrorResponse("Failed to open file"))
	}
	defer fileContent.Close()

	response, err := api.service.ParseAndStoreCSV(fileContent)
	if err != nil {
		return ctx.Status(400).JSON(dto.CreateErrorResponse(err.Error()))
	}

	return ctx.JSON(dto.CreateSuccessResponse("Statement uploaded successfully", response))
}

func (api *transactionHandler) GetBalance(ctx *fiber.Ctx) error {
	response, err := api.service.CalculateBalance()
	if err != nil {
		return ctx.Status(500).JSON(dto.CreateErrorResponse(err.Error()))
	}

	return ctx.JSON(dto.CreateSuccessResponse("Balance calculated successfully", response))
}

func (api *transactionHandler) GetIssues(ctx *fiber.Ctx) error {
	var pagination dto_transaction.PaginationDTO
	var sorting dto_transaction.SortingDTO

	if err := ctx.QueryParser(&pagination); err != nil {
		return ctx.Status(400).JSON(dto.CreateErrorResponse("Invalid query parameters"))
	}

	if err := ctx.QueryParser(&sorting); err != nil {
		return ctx.Status(400).JSON(dto.CreateErrorResponse("Invalid query parameters"))
	}

	response, err := api.service.GetIssues(pagination, sorting)
	if err != nil {
		return ctx.Status(500).JSON(dto.CreateErrorResponse(err.Error()))
	}

	return ctx.JSON(dto.CreateSuccessResponse("Issues retrieved successfully", response))
}
