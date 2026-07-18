package app

import (
	"fmt"

	"github.com/Onubee/task/internal/config"
)

type App struct {
	config *config.Config
}

// NewApp — создает новое приложение
func NewApp(cfg *config.Config) *App {
	return &App{
		config: cfg,
	}
}

// Run — запускает приложение
func (a *App) Run() error {
	fmt.Println("🚀 Application started!")
	fmt.Printf("📦 Config: DB=%s:%d, Server=:%d\n",
		a.config.DBHost,
		a.config.DBPort,
		a.config.ServerPort,
	)
	return nil
}
