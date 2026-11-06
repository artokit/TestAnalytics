package response

type CurrencyResponse struct {
	Base   string             `json:"base"`
	Prices map[string]float64 `json:"prices"`
}

type CoinGeckoRates struct {
	Tether CoinGeckoPrice `json:"tether"`
	Solana CoinGeckoPrice `json:"solana"`
}

type CoinGeckoPrice struct {
	Rub float64 `json:"rub"`
}
