package orders

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// validate the data before creating an order
func ValidateAddOrder(order *Order) error {

	if order == nil {
		return fmt.Errorf("%w: order cannot be nil", ErrInvalidOrderData)
	}

	if order.BusinessUserID <= 0 {
		return fmt.Errorf("%w: invalid business user ID", ErrInvalidOrderData)
	}

	// Required fields validation
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

	// Date validation
	if order.PreferredDeliveryAt.IsZero() {
		return fmt.Errorf("%w: preferred delivery time is required", ErrInvalidOrderData)
	}

	// Price validation
	if order.ItemPrice < 0 || order.DeliveryFee < 0 {
		return fmt.Errorf("%w: prices cannot be negative", ErrInvalidOrderData)
	}

	// Package weight validation
	if order.PackageWeightKg != nil && *order.PackageWeightKg <= 0 {
		return fmt.Errorf("%w: package weight must be positive", ErrInvalidOrderData)
	}

	// Default values
	if order.DeliveryFeePayer == "" {
		order.DeliveryFeePayer = "customer"
	}

	if order.PaymentType == "" {
		order.PaymentType = "cod"
	}

	// Enum validation
	if order.DeliveryFeePayer != "customer" && order.DeliveryFeePayer != "business" {
		return fmt.Errorf("%w: invalid delivery fee payer", ErrInvalidOrderData)
	}

	if order.PaymentType != "cod" && order.PaymentType != "prepaid" {
		return fmt.Errorf("%w: invalid payment type", ErrInvalidOrderData)
	}

	// String length validation
	if utf8.RuneCountInString(order.TrackingNumber) > 30 {
		return fmt.Errorf("%w: tracking number must not exceed 30 characters", ErrInvalidOrderData)
	}

	if utf8.RuneCountInString(order.CustomerName) > 100 {
		return fmt.Errorf("%w: customer name must not exceed 100 characters", ErrInvalidOrderData)
	}

	if utf8.RuneCountInString(order.CustomerPhone) > 20 {
		return fmt.Errorf("%w: customer phone must not exceed 20 characters", ErrInvalidOrderData)
	}

	return nil
}
