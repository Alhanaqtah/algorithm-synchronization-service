package models

import "time"

// Client represents information about a client.
type Client struct {
	ID          int64     `json:"id,omitempty" example:"1"`
	ClientName  string    `json:"client_name,omitempty" example:"Client A"`
	Version     int       `json:"version,omitempty" example:"1"`
	Image       string    `json:"image,omitempty" example:"client-image:latest"`
	CPU         string    `json:"cpu,omitempty" example:"2 cores"`
	Memory      string    `json:"memory,omitempty" example:"4 GB"`
	Priority    float64   `json:"priority,omitempty" example:"0.75"`
	NeedRestart *bool     `json:"need_restart,omitempty" example:"false"`
	SpawnedAt   time.Time `json:"spawned_at,omitempty" example:"2024-07-17T12:00:00Z"`
	CreatedAt   time.Time `json:"created_at,omitempty" example:"2024-07-01T08:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at,omitempty" example:"2024-07-17T14:30:00Z"`
}
