package orders

import "time"

type Order struct {
	ID                  int64      `json:"id"`
	TrackingNumber      string     `json:"tracking_number"`
	BusinessUserID      int64      `json:"business_user_id"`
	CustomerName        string     `json:"customer_name"`
	CustomerPhone       string     `json:"customer_phone"`
	PickupAddress       string     `json:"pickup_address"`
	DropoffAddress      string     `json:"dropoff_address"`
	DeliveryNotes       *string    `json:"delivery_notes"`
	PackageDescription  string     `json:"package_description"`
	PackageWeightKg     *float64   `json:"package_weight_kg"`
	PreferredDeliveryAt time.Time  `json:"preferred_delivery_at"`
	ItemPrice           float64    `json:"item_price"`
	DeliveryFee         float64    `json:"delivery_fee"`
	DeliveryFeePayer    string     `json:"delivery_fee_payer"`
	PaymentType         string     `json:"payment_type"`
	TotalCustomerPrice  float64    `json:"total_customer_price"`
	CODAmount           float64    `json:"cod_amount"`
	CODCollectedAt      *time.Time `json:"cod_collected_at"`
	Status              string     `json:"status"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	CourierName         *string    `json:"courier_name"`
	CourierPhone        *string    `json:"courier_phone"`
}
