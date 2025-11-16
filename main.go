package main

import (
	"log"
	"time"

	"firstpersoncode/go-uploader/internal/config"
	"firstpersoncode/go-uploader/internal/middlewares"
	"firstpersoncode/go-uploader/internal/modules/auth"
	"firstpersoncode/go-uploader/internal/modules/transaction"
	"firstpersoncode/go-uploader/internal/repositories"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func main() {

	app := fiber.New()

	app.Use(limiter.New(limiter.Config{
		Max:               20,
		Expiration:        30 * time.Second,
		LimiterMiddleware: limiter.SlidingWindow{},
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	userRepo := repositories.NewUserRepository()
	sessionRepo := repositories.NewSessionRepository()
	transactionRepo := repositories.NewTransactionRepository()

	sessionMiddleware := middlewares.NewSessionMiddleware(sessionRepo)

	authService := auth.NewAuthService(userRepo, sessionRepo)
	authHandler := auth.NewAuthHandler(authService)

	app.Post("/signup", authHandler.SignUp)
	app.Post("/signin", authHandler.SignIn)
	app.Post("/signout", authHandler.SignOut)
	app.Get("/session", sessionMiddleware.Handle, authHandler.Session)

	transactionService := transaction.NewTransactionService(transactionRepo)
	transactionHandler := transaction.NewTransactionHandler(transactionService)

	app.Post("/upload", sessionMiddleware.Handle, transactionHandler.UploadStatement)
	app.Get("/balance", sessionMiddleware.Handle, transactionHandler.GetBalance)
	app.Get("/issues", sessionMiddleware.Handle, transactionHandler.GetIssues)

	configServer := config.Get().Server

	host := configServer.Host
	port := configServer.Port

	log.Printf("Server running on %s:%s", host, port)
	log.Fatal(app.Listen(host + ":" + port))
}
