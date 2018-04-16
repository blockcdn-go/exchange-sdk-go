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

var events = []event{{
	Event: "addChannel",
	Parameters: parameter{
		Base:    "okb",
		Binary:  "1",
		Product: "spot",
		Quote:   "btc",
		Type:    "ticker",
	},
}, {
	Event: "addChannel",
	Parameters: parameter{
		Base:    "ont",
		Binary:  "1",
		Product: "spot",
		Quote:   "btc",
		Type:    "ticker",
	},
}, {
	Event: "addChannel",
	Parameters: parameter{
		Base:    "enj",
		Binary:  "1",
		Product: "spot",
		Quote:   "btc",
		Type:    "ticker",
	},
}, {
	Event: "addChannel",
	Parameters: parameter{
		Base:    "dadi",
		Binary:  "1",
		Product: "spot",
		Quote:   "btc",
		Type:    "ticker",
	},
}, {
	Event: "addChannel",
	Parameters: parameter{
		Base:    "wfee",
		Binary:  "1",
		Product: "spot",
		Quote:   "btc",
		Type:    "ticker",
	},
}, {
	Event: "addChannel",
	Parameters: parameter{
		Base:    "ren",
		Binary:  "1",
		Product: "spot",
		Quote:   "btc",
		Type:    "ticker",
	},
}, {
	Event: "addChannel",
	Parameters: parameter{
		Base:    "tra",
		Binary:  "1",
		Product: "spot",
		Quote:   "btc",
		Type:    "ticker",
	},
}, {
	Event: "addChannel",
	Parameters: parameter{
		Base:    "trio",
		Binary:  "1",
		Product: "spot",
		Quote:   "btc",
		Type:    "ticker",
	},
}, {
	Event: "addChannel",
	Parameters: parameter{
		Base:    "rfr",
		Binary:  "1",
		Product: "spot",
		Quote:   "btc",
		Type:    "ticker",
	},
}, {
	Event: "addChannel",
	Parameters: parameter{
		Base:    "gsc",
		Binary:  "1",
		Product: "spot",
		Quote:   "btc",
		Type:    "ticker",
	},
}}
