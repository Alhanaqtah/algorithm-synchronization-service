package models

type AlgorithmStatus struct {
	ClientID int  `json:"client_id,omitempty"`
	VWAP     bool `json:"vwap,omitempty"`
	TWAP     bool `json:"twap,omitempty"`
	HFT      bool `json:"hft,omitempty"`
}
