package entity

type Aggregation struct {
	Dex map[string]Platform `json:"dex"`
}

type Platform struct {
	Name     string  `json:"name"`
	Address  Address `json:"address"`
	Fee      float64 `json:"fee"`
	Reserve0 float64 `json:"reserve0"`
	Reserve1 float64 `json:"reserve1"`
	IsActive bool    `json:"is_active"`
	Status   string  `json:"status"`
	Balance  string  `json:"balance"`
	Price    float64 `json:"-"`
	NewPrice float64 `json:"-"`
	Dx       float64 `json:"-"`
	Dy       float64 `json:"-"`
}

type Address struct {
	Bounce   string `json:"bounce"`
	UnBounce string `json:"unbounce"`
}
