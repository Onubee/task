package command

import (
	"context"
	"sync"
	"time"

	"github.com/Onubee/task/internal/domain"
	"github.com/Onubee/task/internal/logger"
	"github.com/Onubee/task/internal/repository"
	"github.com/Onubee/task/internal/service"
	"github.com/Onubee/task/pkg/httpclient"

	"golang.org/x/time/rate"
)

type DownloadStats struct {
	TotalPages      int
	SuccessfulPages int
	FailedPages     int
	TotalProducts   int
	TotalClients    int
	mu              sync.Mutex
}

func (s *DownloadStats) IncrementProducts(count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalProducts += count
}

func (s *DownloadStats) IncrementClients(count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalClients += count
}

type DownloadCommand struct {
	productRepo   repository.ProductRepository
	clientRepo    repository.ClientRepository
	taskRepo      repository.TaskRepository
	sources       []string
	clientsURL    string
	timeout       time.Duration
	maxRetries    int
	retryDelay    time.Duration
	pageStart     int
	pageLimit     int
	maxPages      int
	maxConcurrent int
	rateLimiter   *rate.Limiter
	requestDelay  time.Duration
	logger        *logger.Logger
	httpClient    *httpclient.Client
	workerPool    chan struct{}
	errorCount    int
	errorMutex    sync.Mutex
	stats         *DownloadStats
	mu            sync.RWMutex
	currentTask   *domain.Task
	isRunning     bool
}

func NewDownloadCommand(
	pRepo repository.ProductRepository,
	cRepo repository.ClientRepository,
	tRepo repository.TaskRepository,
	sources []string,
	clientsURL string,
	timeout time.Duration,
	maxRetries int,
	retryDelay time.Duration,
	pageStart int,
	pageLimit int,
	maxPages int,
	maxConcurrent int,
	rateLimit int,
	burstSize int,
	requestDelay time.Duration,
	logger *logger.Logger,
	httpClient *httpclient.Client,
) *DownloadCommand {
	return &DownloadCommand{
		productRepo:   pRepo,
		clientRepo:    cRepo,
		taskRepo:      tRepo,
		sources:       sources,
		clientsURL:    clientsURL,
		timeout:       timeout,
		maxRetries:    maxRetries,
		retryDelay:    retryDelay,
		pageStart:     pageStart,
		pageLimit:     pageLimit,
		maxPages:      maxPages,
		maxConcurrent: maxConcurrent,
		rateLimiter:   rate.NewLimiter(rate.Limit(rateLimit), burstSize),
		requestDelay:  requestDelay,
		logger:        logger,
		httpClient:    httpClient,
		workerPool:    make(chan struct{}, maxConcurrent),
		stats:         &DownloadStats{},
	}
}

func (d *DownloadCommand) Download(ctx context.Context) error {
	d.mu.Lock()
	if d.isRunning {
		d.mu.Unlock()
		return service.ErrDownloadInProgress
	}
	d.isRunning = true
	d.mu.Unlock()

	defer func() {
		d.mu.Lock()
		d.isRunning = false
		d.currentTask = nil
		d.mu.Unlock()
	}()

	return d.doDownload(ctx)
}

func (d *DownloadCommand) GetStatus() (*domain.Task, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.currentTask != nil {
		return d.currentTask, nil
	}
	return d.taskRepo.GetActive()
}

func (d *DownloadCommand) Cancel(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.isRunning || d.currentTask == nil {
		return service.ErrDownloadInProgress
	}
	return d.taskRepo.Finish(d.currentTask.ID, "cancelled")
}
