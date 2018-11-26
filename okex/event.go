package okex

type Event struct {
	Event      string    `json:"event"`
	Parameters Parameter `json:"parameters"`
}

type Parameter struct {
	Base    string `json:"base"`
	Binary  string `json:"binary"`
	Product string `json:"product"`
	Quote   string `json:"quote"`
	Type    string `json:"type"`
}

var _events = []Event{{
	Event: "addChannel",
	Parameters: Parameter{
		Base:    "okb",
		Binary:  "1",
		Product: "spot",
		Quote:   "btc",
		Type:    "ticker",
	},
}, {
	Event: "addChannel",
	Parameters: Parameter{
		Base:    "ont",
		Binary:  "1",
		Product: "spot",
		Quote:   "btc",
		Type:    "ticker",
	},
}, {
	Event: "addChannel",
	Parameters: Parameter{
		Base:    "enj",
		Binary:  "1",
		Product: "spot",
		Quote:   "btc",
		Type:    "ticker",
	},
}, {
	Event: "addChannel",
	Parameters: Parameter{
		Base:    "dadi",
		Binary:  "1",
		Product: "spot",
		Quote:   "btc",
		Type:    "ticker",
	},
}, {
	Event: "addChannel",
	Parameters: Parameter{
		Base:    "wfee",
		Binary:  "1",
		Product: "spot",
		Quote:   "btc",
		Type:    "ticker",
	},
}, {
	Event: "addChannel",
	Parameters: Parameter{
		Base:    "ren",
		Binary:  "1",
		Product: "spot",
		Quote:   "btc",
		Type:    "ticker",
	},
}, {
	Event: "addChannel",
	Parameters: Parameter{
		Base:    "tra",
		Binary:  "1",
		Product: "spot",
		Quote:   "btc",
		Type:    "ticker",
	},
}, {
	Event: "addChannel",
	Parameters: Parameter{
		Base:    "trio",
		Binary:  "1",
		Product: "spot",
		Quote:   "btc",
		Type:    "ticker",
	},
}, {
	Event: "addChannel",
	Parameters: Parameter{
		Base:    "rfr",
		Binary:  "1",
		Product: "spot",
		Quote:   "btc",
		Type:    "ticker",
	},
}, {
	Event: "addChannel",
	Parameters: Parameter{
		Base:    "gsc",
		Binary:  "1",
		Product: "spot",
		Quote:   "btc",
		Type:    "ticker",
	},
}}
