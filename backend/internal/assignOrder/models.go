package assignOrder

import "time"

// Assignment represents a courier assignment.
type Assignment struct {
	ID          int        `json:"id"`
	OrderID     int        `json:"order_id"`
	CourierID   int        `json:"courier_id"`
	AssignedBy  int        `json:"assigned_by"`
	Status      string     `json:"status"`
	AssignedAt  time.Time  `json:"assigned_at"`
	RespondedAt *time.Time `json:"responded_at,omitempty"`
}

// Courier
type Courier struct {
    ID       int64  `json:"id"`
    FullName string `json:"full_name"`
}