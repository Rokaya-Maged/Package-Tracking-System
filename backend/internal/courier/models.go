
package courier

import "time"

// Assignment statuses: courier_assignments.status
const (
	AssignmentPending    = "pending"
	AssignmentAccepted   = "accepted"
	AssignmentDeclined   = "declined"
	AssignmentReassigned = "reassigned"
)

// Shipment statuses: orders.status
const (
	OrderPending   = "pending"
	OrderAssigned  = "assigned"
	OrderPickedUp  = "picked_up"
	OrderInTransit = "in_transit"
	OrderDelivered = "delivered"
	OrderCancelled = "cancelled"
)

// Assignment represents a courier assignment.
type Assignment struct {
	ID         int64      `json:"id"`
	OrderID    int64      `json:"order_id"`
	CourierID  int64      `json:"courier_id"`
	AssignedBy int64      `json:"assigned_by"`
	Status     string     `json:"status"`
	AssignedAt time.Time  `json:"assigned_at"`
	RespondedAt *time.Time `json:"responded_at,omitempty"`
}

// CourierAssignedOrder is the shipment information
// visible to the courier.
type CourierAssignedOrder struct {
	OrderID             int64      `json:"order_id"`
	TrackingNumber      string     `json:"tracking_number"`
	CustomerName        string     `json:"customer_name"`
	CustomerPhone       string     `json:"customer_phone"`
	PickupAddress       string     `json:"pickup_address"`
	DropoffAddress      string     `json:"dropoff_address"`
	DeliveryNotes       *string    `json:"delivery_notes,omitempty"`
	PackageDescription  string     `json:"package_description"`
	PackageWeightKG     *float64   `json:"package_weight_kg,omitempty"`
	PaymentType         string     `json:"payment_type"`
	CODAmount           float64    `json:"cod_amount"`
	DeliveryFee         float64    `json:"delivery_fee"`
	PreferredDeliveryAt time.Time  `json:"preferred_delivery_at"`
	OrderStatus         string     `json:"order_status"`

	AssignmentID     int64     `json:"assignment_id"`
	AssignmentStatus string    `json:"assignment_status"`
	AssignedAt       time.Time `json:"assigned_at"`
}

// CourierStatusUpdateRequest is sent when a courier
// updates the delivery progress.
type CourierStatusUpdateRequest struct {
	Status string `json:"status"`
	Notes  string `json:"notes"`
}

// APIResponse provides a consistent response structure.
type APIResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}
