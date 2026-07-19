package command

import (
	"context"
	"fmt"
	"time"
)

func (d *DownloadCommand) doDownload(ctx context.Context) error {
	startTime := time.Now()
	d.logger.Info("🚀 Starting download process...")

	activeTask, err := d.taskRepo.GetActive()
	if err != nil {
		return fmt.Errorf("failed to check active task: %w", err)
	}

	if activeTask != nil {
		d.logger.Warn("⚠️  Stopping previous task (ID: %d)", activeTask.ID)
		if err := d.taskRepo.Finish(activeTask.ID, "stopped"); err != nil {
			d.logger.Warn("⚠️  Failed to stop previous task: %v", err)
		}
	}

	task, err := d.taskRepo.Create()
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	d.currentTask = task
	d.logger.Info("✅ Task created (ID: %d)", task.ID)

	defer func() {
		if r := recover(); r != nil {
			d.logger.Error("🔥 Panic recovered: %v", r)
			_ = d.taskRepo.Finish(task.ID, "stopped")
		}
	}()

	if err := d.downloadWithTimeout(ctx); err != nil {
		d.logger.Error("❌ Download failed: %v", err)
		_ = d.taskRepo.Finish(task.ID, "stopped")
		return err
	}

	elapsed := time.Since(startTime)
	finishStatus := "completed"
	if d.errorCount > 0 {
		finishStatus = "completed_with_errors"
	}

	if err := d.taskRepo.Finish(task.ID, finishStatus); err != nil {
		d.logger.Warn("⚠️  Failed to finish task: %v", err)
	}

	d.logger.Info("✅ Download completed in %v", elapsed)
	d.logger.Info("📊 Statistics:")
	d.logger.Info("  📄 Total pages: %d", d.stats.TotalPages)
	d.logger.Info("  ✅ Successful: %d", d.stats.SuccessfulPages)
	d.logger.Info("  ❌ Failed: %d", d.stats.FailedPages)
	d.logger.Info("  📦 Products: %d", d.stats.TotalProducts)
	d.logger.Info("  👤 Clients: %d", d.stats.TotalClients)
	d.logger.Info("  ⚠️  Errors: %d", d.errorCount)

	if d.errorCount > 0 {
		return fmt.Errorf("download completed with %d errors", d.errorCount)
	}
	return nil
}
