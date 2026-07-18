package app

import (
	"github.com/Onubee/task/internal/config"
	"github.com/Onubee/task/internal/logger"
)

type App struct {
	config *config.Config
	logger *logger.Logger // ← добавили логгер
}

// NewApp — создает новое приложение
func NewApp(cfg *config.Config) *App {
	// Создаем логгер с уровнем из конфига
	log := logger.NewLogger(cfg.LogLevel)

	return &App{
		config: cfg,
		logger: log,
	}
}

// Run — запускает приложение
func (a *App) Run() error {
	a.logger.Info("🚀 Application started!")
	a.logger.Info("📦 Config: DB=%s:%d, Server=:%d",
		a.config.DBHost,
		a.config.DBPort,
		a.config.ServerPort,
	)
	a.logger.Debug("🔍 Product sources: %v", a.config.GetProductURLs())

	return nil
}
