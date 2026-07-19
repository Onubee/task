package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"time"

	"github.com/Onubee/task/internal/config"
	"github.com/Onubee/task/internal/logger"

	_ "github.com/lib/pq"
)

type App struct {
	config *config.Config
	db     *sql.DB
	logger *logger.Logger
	server *http.Server
}

func NewApp(cfg *config.Config) (*App, error) {
	app := &App{
		config: cfg,
		logger: logger.NewLogger(cfg.LogLevel),
	}

	app.logger.Info("🚀 Initializing application...")

	app.logger.Debug("✅ HTTP client created")

	// Инициализация БД
	if err := app.initDB(); err != nil {
		return nil, fmt.Errorf("init DB: %w", err)
	}

	app.logger.Debug("✅ Repositories created")

	// Получаем список URL из конфига
	productSources := cfg.GetProductURLs()
	app.logger.Info("📦 Product sources: %v", productSources)

	app.logger.Debug("✅ Download command created")

	app.logger.Info("✅ Application initialized (handlers pending)")

	return app, nil
}

func (a *App) initDB() error {
	db, err := sql.Open("postgres", a.config.GetDSN())
	if err != nil {
		return fmt.Errorf("open connection: %w", err)
	}

	// Если произойдет ошибка — закрываем соединение
	defer func() {
		if err != nil && db != nil {
			if closeErr := db.Close(); closeErr != nil {
				a.logger.Warn("⚠️  Failed to close database on error: %v", closeErr)
			}
		}
	}()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	a.db = db
	a.logger.Info("✅ Database connected successfully")
	return nil
}

func (a *App) setupRouter() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"ok","message":"service is running"}`)); err != nil {
			// Логируем ошибку, но не падаем
			a.logger.Error("❌ Failed to write response: %v", err)
		}
	})

	a.server = &http.Server{
		Addr:         a.config.GetServerAddress(),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func (a *App) Run() error {
	a.setupRouter()

	go func() {
		a.logger.Info("🚀 Server started on %s", a.server.Addr)
		a.logger.Info("❤️  Health: GET http://localhost:%d/health", a.config.ServerPort)
		a.logger.Info("📥 Download: POST http://localhost:%d/download (coming soon)", a.config.ServerPort)
		a.logger.Info("📊 Stats: GET http://localhost:%d/stats (coming soon)", a.config.ServerPort)

		if err := a.server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				a.logger.Error("❌ Server error: %v", err)
			}
		}
	}()

	return a.waitForShutdown()
}

func (a *App) waitForShutdown() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.logger.Info("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	defer func() {
		if err := a.db.Close(); err != nil {
			a.logger.Error("❌ Failed to close database: %v", err)
		} else {
			a.logger.Info("✅ Database closed")
		}
	}()

	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	a.logger.Info("✅ Server shutdown gracefully")
	return nil
}
