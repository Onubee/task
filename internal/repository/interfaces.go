package repository

import "github.com/Onubee/task/internal/domain"

type ProductRepository interface {
	GetOrCreateBrand(name string) (int, error)
	GetOrCreateCategory(name string) (int, error)
	UpsertProduct(p *domain.NormalizedProduct) error
	GetStats() (*domain.ProductStats, error)
}

type ClientRepository interface {
	UpsertClient(c *domain.Client) error
	UpdateClientProducts(clientID int, productIDs []int) error
}

type TaskRepository interface {
	Create() (*domain.Task, error)
	Finish(id int, status string) error
	GetActive() (*domain.Task, error)
}
