package main

import (
	"log"

	"github.com/Onubee/task/internal/app"
	"github.com/Onubee/task/internal/config"
)

func main() {
	// 1. Загружаем конфиг
	cfg := config.Load()
	log.Printf("✅ Config loaded: %+v", cfg)

	// 2. Создаем приложение (пока заглушка)
	application := app.NewApp(cfg)

	// 3. Запускаем
	if err := application.Run(); err != nil {
		log.Fatalf("❌ App error: %v", err)
	}
}
