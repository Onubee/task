package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Onubee/task/internal/domain"
)

type ProductRepositoryImpl struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepositoryImpl {
	return &ProductRepositoryImpl{db: db}
}

func (r *ProductRepositoryImpl) GetOrCreateBrand(name string) (int, error) {
	if name == "" {
		return 0, errors.New("brand name cannot be empty")
	}

	var id int
	err := r.db.QueryRow(`SELECT id FROM brands WHERE name = $1`, name).Scan(&id)
	if err == sql.ErrNoRows {
		err = r.db.QueryRow(`INSERT INTO brands (name) VALUES ($1) RETURNING id`, name).Scan(&id)
		if err != nil {
			return 0, fmt.Errorf("failed to create brand: %w", err)
		}
		return id, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get brand: %w", err)
	}
	return id, nil
}

func (r *ProductRepositoryImpl) GetOrCreateCategory(name string) (int, error) {
	if name == "" {
		return 0, errors.New("category name cannot be empty")
	}

	var id int
	err := r.db.QueryRow(`SELECT id FROM categories WHERE name = $1`, name).Scan(&id)
	if err == sql.ErrNoRows {
		err = r.db.QueryRow(`INSERT INTO categories (name) VALUES ($1) RETURNING id`, name).Scan(&id)
		if err != nil {
			return 0, fmt.Errorf("failed to create category: %w", err)
		}
		return id, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get category: %w", err)
	}
	return id, nil
}

func (r *ProductRepositoryImpl) UpsertProduct(p *domain.NormalizedProduct) error {
	if p == nil {
		return errors.New("product cannot be nil")
	}
	if p.ID <= 0 {
		return errors.New("product ID must be positive")
	}
	if p.Name == "" {
		return errors.New("product name cannot be empty")
	}
	if p.BrandID <= 0 {
		return errors.New("invalid brand ID")
	}
	if p.CategoryID <= 0 {
		return errors.New("invalid category ID")
	}
	if p.Price < 0 {
		return errors.New("price cannot be negative")
	}
	if p.Stock < 0 {
		return errors.New("stock cannot be negative")
	}

	query := `
		INSERT INTO products (id, name, brand_id, category_id, price, stock, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			brand_id = EXCLUDED.brand_id,
			category_id = EXCLUDED.category_id,
			price = EXCLUDED.price,
			stock = EXCLUDED.stock,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := r.db.Exec(query, p.ID, p.Name, p.BrandID, p.CategoryID, p.Price, p.Stock)
	if err != nil {
		return fmt.Errorf("failed to upsert product: %w", err)
	}
	return nil
}

func (r *ProductRepositoryImpl) GetStats() (*domain.ProductStats, error) {
	stats := &domain.ProductStats{}

	err := r.db.QueryRow(`SELECT COUNT(*) FROM products`).Scan(&stats.TotalProducts)
	if err != nil {
		return nil, fmt.Errorf("failed to count products: %w", err)
	}

	err = r.db.QueryRow(`SELECT COUNT(*) FROM clients`).Scan(&stats.TotalClients)
	if err != nil {
		return nil, fmt.Errorf("failed to count clients: %w", err)
	}

	err = r.db.QueryRow(`SELECT COUNT(*) FROM brands`).Scan(&stats.TotalBrands)
	if err != nil {
		return nil, fmt.Errorf("failed to count brands: %w", err)
	}

	err = r.db.QueryRow(`SELECT COUNT(*) FROM categories`).Scan(&stats.TotalCategories)
	if err != nil {
		return nil, fmt.Errorf("failed to count categories: %w", err)
	}

	return stats, nil
}
