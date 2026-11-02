package orders

import "time"

type Order struct {
	ID              int64     `json:"id"`
	CustomerID      int64     `json:"customer_id"`
	PickupAddress   string    `json:"pickup_address"`
	DeliveryAddress string    `json:"delivery_address"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}
