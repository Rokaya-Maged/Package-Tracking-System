
package courier

import "time"

// Assignment statuses belong to courier_assignments.status.
const (
	AssignmentPending    = "pending"
	AssignmentAccepted   = "accepted"
	AssignmentDeclined   = "declined"
	AssignmentReassigned = "reassigned"
)

// Order statuses belong to orders.status.
const (
	OrderPending   = "pending"
	OrderAssigned  = "assigned"
	OrderPickedUp  = "picked_up"
	OrderInTransit = "in_transit"
	OrderDelivered = "delivered"
	OrderCancelled = "cancelled"
)

// Assignment represents a courier's assignment to an order.
type Assignment struct {
	ID         int64      `json:"id"`
	OrderID    int64      `json:"order_id"`
	CourierID  int64      `json:"courier_id"`
	AssignedBy int64      `json:"assigned_by"`
	Status     string     `json:"status"`
	AssignedAt time.Time  `json:"assigned_at"`
	RespondedAt *time.Time `json:"responded_at,omitempty"`
}

// CourierAssignedOrder contains the order information
// required by the courier dashboard and order details page.
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
	PreferredDeliveryAt time.Time  `json:"preferred_delivery_at"`

	ItemPrice           float64    `json:"item_price"`
	DeliveryFee         float64    `json:"delivery_fee"`
	DeliveryFeePayer    string     `json:"delivery_fee_payer"`
	PaymentType         string     `json:"payment_type"`
	TotalCustomerPrice  float64    `json:"total_customer_price"`
	CODAmount           float64    `json:"cod_amount"`
	CODCollectedAt      *time.Time `json:"cod_collected_at,omitempty"`

	OrderStatus         string     `json:"order_status"`
	AssignmentID        int64      `json:"assignment_id"`
	AssignmentStatus    string     `json:"assignment_status"`
	AssignedAt          time.Time  `json:"assigned_at"`
}

// CourierStatusUpdateRequest is sent when the courier
// updates the delivery status.
type CourierStatusUpdateRequest struct {
	Status string `json:"status"`
	Notes  string `json:"notes"`
}

// APIResponse provides a consistent API response structure.
type APIResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// StatusHistory represents one recorded order status change.
type StatusHistory struct {
	ID        int64     `json:"id"`
	OrderID   int64     `json:"order_id"`
	OldStatus *string   `json:"old_status,omitempty"`
	NewStatus string    `json:"new_status"`
	ChangedBy int64     `json:"changed_by"`
	Notes     *string   `json:"notes,omitempty"`
	ChangedAt time.Time `json:"changed_at"`
}
