package orders

import (
	"fmt"
	"strings"
)

// validate the data before creating an order
func ValidateAddOrder(order *Order) error {
	if order == nil {
		return fmt.Errorf("%w: order cannot be nil", ErrInvalidOrderData)
	}
	if order.BusinessUserID <= 0 {
		return fmt.Errorf("%w: invalid business user ID", ErrInvalidOrderData)
	}
	for name, value := range map[string]string{
		"tracking number":     order.TrackingNumber,
		"customer name":       order.CustomerName,
		"customer phone":      order.CustomerPhone,
		"pickup address":      order.PickupAddress,
		"dropoff address":     order.DropoffAddress,
		"package description": order.PackageDescription,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%w: %s is required", ErrInvalidOrderData, name)
		}
	}
	if order.PreferredDeliveryAt.IsZero() {
		return fmt.Errorf("%w: preferred delivery time is required", ErrInvalidOrderData)
	}
	if order.ItemPrice < 0 || order.DeliveryFee < 0 {
		return fmt.Errorf("%w: prices cannot be negative", ErrInvalidOrderData)
	}
	if order.PackageWeightKg != nil && *order.PackageWeightKg <= 0 {
		return fmt.Errorf("%w: package weight must be positive", ErrInvalidOrderData)
	}
	if order.DeliveryFeePayer == "" {
		order.DeliveryFeePayer = "customer"
	}
	if order.PaymentType == "" {
		order.PaymentType = "cod"
	}
	if order.DeliveryFeePayer != "customer" && order.DeliveryFeePayer != "business" {
		return fmt.Errorf("%w: invalid delivery fee payer", ErrInvalidOrderData)
	}
	if order.PaymentType != "cod" && order.PaymentType != "prepaid" {
		return fmt.Errorf("%w: invalid payment type", ErrInvalidOrderData)
	}
	return nil
}
