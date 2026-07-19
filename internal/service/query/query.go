package query

import (
	"context"
	"time"

	"github.com/Onubee/task/internal/domain"
	"github.com/Onubee/task/internal/logger"
	"github.com/Onubee/task/internal/repository"
	"github.com/Onubee/task/internal/service"
)

type StatsQuery struct {
	productRepo repository.ProductRepository
	logger      *logger.Logger
	cache       *StatsCache
}

type StatsCache struct {
	stats     *domain.ProductStats
	timestamp time.Time
	ttl       time.Duration
}

func NewStatsCache(ttl time.Duration) *StatsCache {
	return &StatsCache{ttl: ttl}
}

func (c *StatsCache) Get() (*domain.ProductStats, bool) {
	if c.stats == nil || time.Since(c.timestamp) > c.ttl {
		return nil, false
	}
	return c.stats, true
}

func (c *StatsCache) Set(stats *domain.ProductStats) {
	c.stats = stats
	c.timestamp = time.Now()
}

func (c *StatsCache) Invalidate() {
	c.stats = nil
}

func NewStatsQuery(
	productRepo repository.ProductRepository,
	logger *logger.Logger,
) *StatsQuery {
	return &StatsQuery{
		productRepo: productRepo,
		logger:      logger,
		cache:       NewStatsCache(5 * time.Second),
	}
}

func (q *StatsQuery) GetStats(ctx context.Context) (*domain.ProductStats, error) {
	if stats, ok := q.cache.Get(); ok {
		q.logger.Debug("📊 Stats from cache")
		return stats, nil
	}

	stats, err := q.productRepo.GetStats()
	if err != nil {
		return nil, err
	}

	q.cache.Set(stats)
	q.logger.Debug("📊 Stats from database")
	return stats, nil
}

func (q *StatsQuery) GetHealth(ctx context.Context) (*service.HealthStatus, error) {
	status := &service.HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	stats, err := q.GetStats(ctx)
	if err != nil {
		status.Status = "unhealthy"
		status.Database = "error: " + err.Error()
		return status, nil
	}

	status.Database = "ok"
	status.Stats = stats
	return status, nil
}

func (q *StatsQuery) InvalidateCache() {
	q.cache.Invalidate()
	q.logger.Debug("🗑️ Stats cache invalidated")
}
