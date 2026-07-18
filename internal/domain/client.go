package domain

type Client struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Products  []int  `json:"products"`
}
