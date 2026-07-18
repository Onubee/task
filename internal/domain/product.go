package domain

type Product struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Brand    string `json:"brand"`
	Category string `json:"category"`
	Price    string `json:"price"`
	Stock    int    `json:"stock"`
}

type NormalizedProduct struct {
	ID         int
	Name       string
	BrandID    int
	CategoryID int
	Price      float64
	Stock      int
}

type Brand struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ProductStats struct {
	TotalProducts   int `json:"products"`
	TotalClients    int `json:"clients"`
	TotalBrands     int `json:"brands"`
	TotalCategories int `json:"categories"`
}
