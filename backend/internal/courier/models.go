package courier

import "time"

type Assignment struct {
	ID          int64      `json:"id"`
	OrderID     int64      `json:"order_id"`
	CourierID   int64      `json:"courier_id"`
	Status      string     `json:"status"`
	AssignedAt  time.Time  `json:"assigned_at"`
	RespondedAt *time.Time `json:"responded_at,omitempty"`
}

type AssignmentResponse struct {
	Accepted bool `json:"accepted"`
}
