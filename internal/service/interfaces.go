package service

import (
	"context"

	"github.com/Onubee/task/internal/domain"
)

type CommandService interface {
	Download(ctx context.Context) error
	GetStatus() (*domain.Task, error)
	Cancel(ctx context.Context) error
}

type QueryService interface {
	GetStats(ctx context.Context) (*domain.ProductStats, error)
	GetHealth(ctx context.Context) (*HealthStatus, error)
}

type HealthStatus struct {
	Status    string               `json:"status"`
	Database  string               `json:"database"`
	Timestamp string               `json:"timestamp"`
	Stats     *domain.ProductStats `json:"stats,omitempty"`
}
