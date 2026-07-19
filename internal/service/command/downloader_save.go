package command

import (
	"fmt"

	"github.com/Onubee/task/internal/domain"
	"github.com/Onubee/task/internal/service"
)

func (d *DownloadCommand) saveProductWithRecovery(raw domain.Product) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return d.saveProduct(raw)
}

func (d *DownloadCommand) saveProduct(raw domain.Product) error {
	price, err := service.NormalizePrice(raw.Price)
	if err != nil {
		return fmt.Errorf("normalize price: %w", err)
	}

	brandID, err := d.productRepo.GetOrCreateBrand(raw.Brand)
	if err != nil {
		return fmt.Errorf("get/create brand: %w", err)
	}

	catID, err := d.productRepo.GetOrCreateCategory(raw.Category)
	if err != nil {
		return fmt.Errorf("get/create category: %w", err)
	}

	norm := &domain.NormalizedProduct{
		ID:         raw.ID,
		Name:       raw.Name,
		BrandID:    brandID,
		CategoryID: catID,
		Price:      price,
		Stock:      raw.Stock,
	}

	return d.productRepo.UpsertProduct(norm)
}

func (d *DownloadCommand) saveClientWithRecovery(raw domain.Client) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return d.saveClient(raw)
}

func (d *DownloadCommand) saveClient(raw domain.Client) error {
	if err := d.clientRepo.UpsertClient(&raw); err != nil {
		return err
	}
	return d.clientRepo.UpdateClientProducts(raw.ID, raw.Products)
}
