package repository

import (
	"database/sql"
	"time"

	"github.com/Onubee/task/internal/domain"
)

type TaskRepositoryImpl struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepositoryImpl {
	return &TaskRepositoryImpl{db: db}
}

func (r *TaskRepositoryImpl) Create() (*domain.Task, error) {
	var task domain.Task
	query := `INSERT INTO tasks (status) VALUES ('in_progress') RETURNING id, started_at`
	err := r.db.QueryRow(query).Scan(&task.ID, &task.StartedAt)
	if err != nil {
		return nil, err
	}
	task.Status = "in_progress"
	return &task, nil
}

func (r *TaskRepositoryImpl) Finish(id int, status string) error {
	query := `UPDATE tasks SET finished_at = $1, status = $2 WHERE id = $3`
	_, err := r.db.Exec(query, time.Now(), status, id)
	return err
}

func (r *TaskRepositoryImpl) GetActive() (*domain.Task, error) {
	var task domain.Task
	query := `SELECT id, started_at, finished_at, status FROM tasks WHERE status = 'in_progress' ORDER BY id DESC LIMIT 1`
	err := r.db.QueryRow(query).Scan(&task.ID, &task.StartedAt, &task.FinishedAt, &task.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}
