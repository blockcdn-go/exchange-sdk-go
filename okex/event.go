package okex

type event struct {
	Event      string    `json:"event"`
	Parameters parameter `json:"parameters"`
}

type parameter struct {
	Base    string `json:"base"`
	Binary  string `json:"binary"`
	Product string `json:"product"`
	Quote   string `json:"quote"`
	Type    string `json:"type"`
}
