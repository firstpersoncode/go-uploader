package middlewares

import (
	"firstpersoncode/go-uploader/domain"
	"firstpersoncode/go-uploader/dto"
	"firstpersoncode/go-uploader/internal/config"
	"firstpersoncode/go-uploader/internal/util"

	"github.com/gofiber/fiber/v2"
)

type SessionMiddleware struct {
	repo domain.SessionRepository
}

func NewSessionMiddleware(repo domain.SessionRepository) *SessionMiddleware {
	return &SessionMiddleware{
		repo: repo,
	}
}

func (s *SessionMiddleware) Handle(ctx *fiber.Ctx) error {
	token := ctx.Cookies(config.Get().App.CookieName)
	if token == "" {
		return ctx.Status(401).JSON(dto.CreateErrorResponse("Unauthorized: No session token"))
	}

	claims, err := util.ValidateJWT(token)
	if err != nil {
		return ctx.Status(401).JSON(dto.CreateErrorResponse(err.Error()))
	}

	session, err := s.repo.FindByID(claims.Sub)
	if err != nil {
		return ctx.Status(401).JSON(dto.CreateErrorResponse(err.Error()))
	}

	ctx.Locals("session", session)

	return ctx.Next()
}
