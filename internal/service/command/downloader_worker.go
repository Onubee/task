package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Onubee/task/internal/domain"
	"github.com/Onubee/task/internal/service"
)

func (d *DownloadCommand) downloadWithTimeout(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, d.timeout*10)
	defer cancel()

	taskChan := make(chan Task, 100)
	resultChan := make(chan TaskResult, 100)
	errorChan := make(chan error, 100)

	var wg sync.WaitGroup
	for i := 0; i < d.maxConcurrent; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			d.worker(ctx, workerID, taskChan, resultChan, errorChan)
		}(i)
	}

	go d.generateTasks(taskChan)

	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	var errList []error
	for {
		select {
		case result, ok := <-resultChan:
			if !ok {
				goto Done
			}
			d.processResult(result, &errList)

		case err, ok := <-errorChan:
			if !ok {
				continue
			}
			if err != nil {
				d.incrementError()
				errList = append(errList, err)
			}

		case <-ctx.Done():
			return fmt.Errorf("download timeout: %w", ctx.Err())
		}
	}

Done:
	if len(errList) > 0 {
		return fmt.Errorf("download completed with %d errors", len(errList))
	}
	return nil
}

type Task struct {
	URL    string
	Type   string
	Page   int
	Result chan *TaskResult
}

type TaskResult struct {
	Products []domain.Product
	Clients  []domain.Client
	Page     int
	Err      error
}

func (d *DownloadCommand) worker(
	ctx context.Context,
	workerID int,
	taskChan <-chan Task,
	resultChan chan<- TaskResult,
	errorChan chan<- error,
) {
	for task := range taskChan {
		select {
		case <-ctx.Done():
			return
		default:
			if err := d.rateLimiter.Wait(ctx); err != nil {
				errorChan <- fmt.Errorf("worker %d: rate limiter: %w", workerID, err)
				continue
			}

			if d.requestDelay > 0 {
				time.Sleep(d.requestDelay)
			}

			result := d.executeTaskWithRecovery(task, workerID)
			resultChan <- result
		}
	}
}

func (d *DownloadCommand) executeTaskWithRecovery(task Task, workerID int) (result TaskResult) {
	defer func() {
		if r := recover(); r != nil {
			d.logger.Error("🔥 Worker %d panic: %v", workerID, r)
			result = TaskResult{
				Page: task.Page,
				Err:  fmt.Errorf("panic: %v", r),
			}
		}
	}()
	return d.executeTask(task)
}

func (d *DownloadCommand) executeTask(task Task) TaskResult {
	pageURL := d.addPaginationParams(task.URL, task.Page, d.pageLimit)
	d.logger.Debug("  📄 Worker fetching %s page %d", task.Type, task.Page)

	var result TaskResult
	result.Page = task.Page

	err := d.loadWithRetry(func() error {
		resp, err := d.httpClient.Get(pageURL)
		if err != nil {
			return fmt.Errorf("HTTP request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("HTTP %d", resp.StatusCode)
		}

		switch task.Type {
		case "products":
			var products []domain.Product
			if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
				return fmt.Errorf("decode: %w", err)
			}
			if len(products) == 0 {
				return service.ErrEmptyPage
			}
			result.Products = products

		case "clients":
			var clients []domain.Client
			if err := json.NewDecoder(resp.Body).Decode(&clients); err != nil {
				return fmt.Errorf("decode: %w", err)
			}
			if len(clients) == 0 {
				return service.ErrEmptyPage
			}
			result.Clients = clients
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, service.ErrEmptyPage) {
			d.logger.Debug("  📄 Page %d is empty, stopping", task.Page)
			return result
		}
		result.Err = err
	} else {
		d.stats.SuccessfulPages++
	}

	d.stats.TotalPages++
	return result
}

func (d *DownloadCommand) processResult(result TaskResult, errList *[]error) {
	if result.Err != nil {
		if !errors.Is(result.Err, service.ErrEmptyPage) {
			d.incrementError()
			*errList = append(*errList, fmt.Errorf("page %d: %w", result.Page, result.Err))
		}
		return
	}

	if len(result.Products) > 0 {
		saved := 0
		for _, p := range result.Products {
			if err := d.saveProductWithRecovery(p); err != nil {
				d.logger.Warn("  ⚠️  Failed to save product %d: %v", p.ID, err)
				d.incrementError()
			} else {
				saved++
			}
		}
		d.stats.IncrementProducts(saved)
		d.logger.Info("  💾 Saved %d/%d products from page %d", saved, len(result.Products), result.Page)
	}

	if len(result.Clients) > 0 {
		saved := 0
		for _, c := range result.Clients {
			if err := d.saveClientWithRecovery(c); err != nil {
				d.logger.Warn("  ⚠️  Failed to save client %d: %v", c.ID, err)
				d.incrementError()
			} else {
				saved++
			}
		}
		d.stats.IncrementClients(saved)
		d.logger.Info("  💾 Saved %d/%d clients from page %d", saved, len(result.Clients), result.Page)
	}
}

func (d *DownloadCommand) loadWithRetry(operation func() error) error {
	var lastErr error
	for attempt := 0; attempt < d.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(d.retryDelay)
		}
		if err := operation(); err != nil {
			lastErr = err
			if errors.Is(err, service.ErrEmptyPage) {
				return err
			}
			continue
		}
		return nil
	}
	return fmt.Errorf("failed after %d retries: %w", d.maxRetries, lastErr)
}

func (d *DownloadCommand) generateTasks(taskChan chan<- Task) {
	defer close(taskChan)

	for _, sourceURL := range d.sources {
		for page := d.pageStart; page <= d.maxPages; page++ {
			taskChan <- Task{
				URL:  sourceURL,
				Type: "products",
				Page: page,
			}
		}
	}

	for page := d.pageStart; page <= d.maxPages; page++ {
		taskChan <- Task{
			URL:  d.clientsURL,
			Type: "clients",
			Page: page,
		}
	}
}

func (d *DownloadCommand) incrementError() {
	d.errorMutex.Lock()
	defer d.errorMutex.Unlock()
	d.errorCount++
}
