package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	App    App
	Server Server
}

func Get() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	return &Config{
		App: App{
			CookieName:     os.Getenv("SESSION_COOKIE_NAME"),
			AllowedOrigins: os.Getenv("ALLOWED_ORIGINS"),
		},
		Server: Server{
			Host: os.Getenv("HOST"),
			Port: os.Getenv("PORT"),
		},
	}
}
