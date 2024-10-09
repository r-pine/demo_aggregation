package entity

type Aggregation struct {
	Dex map[string]Platform `json:"dex"`
}

type Platform struct {
	Address  Address `json:"address"`
	Fee      int64   `json:"fee"`
	Reserve0 int64   `json:"reserve0"`
	Reserve1 int64   `json:"reserve1"`
	IsActive bool    `json:"is_active"`
	Status   string  `json:"status"`
	Balance  string  `json:"balance"`
}

type Address struct {
	Bounce   string `json:"bounce"`
	UnBounce string `json:"unbounce"`
	Raw      string `json:"raw"`
}
