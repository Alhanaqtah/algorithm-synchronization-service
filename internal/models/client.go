package models

import "time"

type Client struct {
	ID          int64     `json:"id,omitempty"`
	ClientName  string    `json:"client_name,omitempty"`
	Version     int       `json:"version,omitempty"`
	Image       string    `json:"image,omitempty"`
	CPU         string    `json:"cpu,omitempty"`
	Memory      string    `json:"memory,omitempty"`
	Priority    float64   `json:"repository,omitempty"`
	NeedRestart *bool     `json:"need_restart,omitempty"`
	SpawnedAt   time.Time `json:"spawned_at,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}
