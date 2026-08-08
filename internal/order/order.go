package order

type Order struct {
	Name     string `json:"name"`
	Product  string `json:"product"`
	Quantity int    `json:"quantity"`
}
