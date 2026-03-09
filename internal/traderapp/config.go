package traderapp

type Config struct {
	//LogFolder        string
	NotifyConfig     NotifyConfig
	MarketData       string
	UseCandleStorage bool
	Brokers          []BrokerConfig    `xml:"Client"`
	Signals          []SignalConfig    `xml:"Signal"`
	Portfolios       []PortfolioConfig `xml:"Portfolio"`
}

type NotifyConfig struct {
	Enabled  bool   `xml:",attr"`
	From     string `xml:",attr"`
	To       string `xml:",attr"`
	Password string `xml:",attr"`
	Host     string `xml:",attr"`
	Port     string `xml:",attr"`
}

type BrokerConfig struct {
	Key  string `xml:",attr"`
	Type string `xml:",attr"`
	Port int    `xml:",attr"`
}

type SignalConfig struct {
	Advisor       string  `xml:",attr"`
	Security      string  `xml:",attr"`
	StdVolatility float64 `xml:",attr"`
	SizeConfig    SizeConfig
}

type SizeConfig struct {
	LongLever  float64 `xml:",attr"`
	ShortLever float64 `xml:",attr"`
	MaxLever   float64 `xml:",attr"`
	Weight     float64 `xml:",attr"`
}

type PortfolioConfig struct {
	Client    string  `xml:",attr"`
	Firm      string  `xml:",attr"`
	Account   string  `xml:"Account,attr"`
	MaxAmount float64 `xml:",attr"` // нельзя задать MaxAmount=0 тк не отличается от None
	Weight    float64 `xml:",attr"` // нельзя задать Weight=0 тк не отличается от None
}
