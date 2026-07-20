package main

import (
	"log"

	"github.com/Onubee/task/internal/app"
	"github.com/Onubee/task/internal/config"
)

func main() {
	cfg := config.Load()

	application, err := app.NewApp(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to create app: %v", err)
	}
	defer func() {
		if err := application.Close(); err != nil {
			log.Printf("⚠️ Failed to close application: %v", err)
		}
	}()

	if err := application.Run(); err != nil {
		log.Fatalf("❌ Failed to run app: %v", err)
	}
}
