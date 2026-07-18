package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/Onubee/task/internal/domain"
)

type ClientRepositoryImpl struct {
	db *sql.DB
}

func NewClientRepository(db *sql.DB) *ClientRepositoryImpl {
	return &ClientRepositoryImpl{db: db}
}

func (r *ClientRepositoryImpl) UpsertClient(c *domain.Client) error {
	if c == nil {
		return errors.New("client cannot be nil")
	}
	if c.ID <= 0 {
		return errors.New("client ID must be positive")
	}
	if c.FirstName == "" {
		return errors.New("client first name cannot be empty")
	}
	if c.LastName == "" {
		return errors.New("client last name cannot be empty")
	}

	query := `
		INSERT INTO clients (id, first_name, last_name)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name
	`

	_, err := r.db.Exec(query, c.ID, c.FirstName, c.LastName)
	if err != nil {
		return fmt.Errorf("failed to upsert client: %w", err)
	}
	return nil
}

func (r *ClientRepositoryImpl) UpdateClientProducts(clientID int, productIDs []int) error {
	if clientID <= 0 {
		return errors.New("invalid client ID")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("⚠️  Failed to rollback: %v", rollbackErr)
			}
		}
	}()

	_, err = tx.Exec(`DELETE FROM client_products WHERE client_id = $1`, clientID)
	if err != nil {
		return fmt.Errorf("failed to delete client products: %w", err)
	}

	if len(productIDs) == 0 {
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit: %w", err)
		}
		return nil
	}

	for _, pid := range productIDs {
		if pid <= 0 {
			continue
		}
		_, err = tx.Exec(
			`INSERT INTO client_products (client_id, product_id) VALUES ($1, $2)`,
			clientID, pid,
		)
		if err != nil {
			log.Printf("⚠️  Failed to insert (%d, %d): %v", clientID, pid, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	return nil
}
