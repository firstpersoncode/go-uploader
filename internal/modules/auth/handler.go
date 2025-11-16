package auth

import (
	"time"

	"firstpersoncode/go-uploader/domain"
	"firstpersoncode/go-uploader/dto"
	dto_session "firstpersoncode/go-uploader/dto/session"

	"github.com/gofiber/fiber/v2"
)

type authHandler struct {
	service domain.SessionService
}

func NewAuthHandler(service domain.SessionService) domain.SessionHandler {
	return &authHandler{
		service: service,
	}
}

func (h *authHandler) SignUp(ctx *fiber.Ctx) error {
	var credentials dto_session.TokenRequestDTO
	if err := ctx.BodyParser(&credentials); err != nil {
		return ctx.Status(400).JSON(dto.CreateErrorResponse("Invalid request body"))
	}

	err := h.service.RegisterUser(&credentials)
	if err != nil {
		return ctx.Status(400).JSON(dto.CreateErrorResponse(err.Error()))
	}

	return ctx.JSON(dto.CreateSuccessResponse("User registered successfully", map[string]interface{}{}))
}

func (h *authHandler) SignIn(ctx *fiber.Ctx) error {
	var credentials dto_session.TokenRequestDTO
	if err := ctx.BodyParser(&credentials); err != nil {
		return ctx.Status(400).JSON(dto.CreateErrorResponse("Invalid request body"))
	}

	token_dto, err := h.service.CreateSession(&credentials)
	if err != nil {
		return ctx.Status(401).JSON(dto.CreateErrorResponse(err.Error()))
	}

	cookie := &fiber.Cookie{
		Name:     "__session__",
		Value:    token_dto.Token,
		Expires:  token_dto.Expiry,
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	}

	ctx.Cookie(cookie)

	return ctx.JSON(dto.CreateSuccessResponse("Signed in successfully", token_dto))
}

func (h *authHandler) SignOut(ctx *fiber.Ctx) error {
	cookie := &fiber.Cookie{
		Name:     "__session__",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	}

	ctx.Cookie(cookie)

	return ctx.JSON(fiber.Map{"message": "Signed out successfully"})
}

func (h *authHandler) Session(ctx *fiber.Ctx) error {
	session_dto, err := h.service.GetUserSession(ctx.Locals("session").(*domain.Session))
	if err != nil {
		return ctx.Status(500).JSON(dto.CreateErrorResponse(err.Error()))
	}

	return ctx.JSON(dto.CreateSuccessResponse("Session retrieved successfully", session_dto))
}

func (h *authHandler) Refresh(ctx *fiber.Ctx) error {
	return ctx.Status(501).JSON(fiber.Map{"error": "Not implemented"})
}
