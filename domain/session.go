package domain

import (
	dto_session "firstpersoncode/go-uploader/dto/session"

	"github.com/gofiber/fiber/v2"
)

type Session struct {
	ID           string
	UserID       string
	RefreshToken string
}

type SessionRepository interface {
	Save(session *Session) (*Session, error)
	FindByID(id string) (*Session, error)
}

type SessionService interface {
	RegisterUser(credentials *dto_session.TokenRequestDTO) error
	CreateSession(credentials *dto_session.TokenRequestDTO) (*dto_session.TokenResponseDTO, error)
	GetUserSession(session *Session) (*dto_session.UserSessionDto, error)
	RefreshToken(refreshToken string) (*dto_session.TokenResponseDTO, error)
}

type SessionHandler interface {
	SignUp(ctx *fiber.Ctx) error
	SignIn(ctx *fiber.Ctx) error
	SignOut(ctx *fiber.Ctx) error
	Session(ctx *fiber.Ctx) error
	Refresh(ctx *fiber.Ctx) error
}
