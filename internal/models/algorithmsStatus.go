package models

// AlgoStatuses represents the status of algorithms for a client.
type AlgoStatuses struct {
	ClientID int   `json:"client_id,omitempty" example:"123"`
	VWAP     *bool `json:"vwap,omitempty" example:"true"`
	TWAP     *bool `json:"twap,omitempty" example:"false"`
	HFT      *bool `json:"hft,omitempty" example:"true"`
}
