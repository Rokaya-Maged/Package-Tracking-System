
package courier

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidAssignment  = errors.New("assignment is not pending")
	ErrInvalidTransition  = errors.New("invalid order status transition")
	ErrNotAccepted        = errors.New("assignment has not been accepted")
	ErrOrderNotAssignable = errors.New("order is not available for this assignment")
	ErrNotFound           = errors.New("assignment or order not found")
)

const MaxStatusNotesLength = 1000

func ValidateCourierID(courierID int64) error {
	if courierID <= 0 {
		return fmt.Errorf("%w: courier ID must be positive", ErrInvalidInput)
	}
	return nil
}

func ValidateOrderID(orderID int64) error {
	if orderID <= 0 {
		return fmt.Errorf("%w: order ID must be positive", ErrInvalidInput)
	}
	return nil
}

func ValidateAssignmentResponse(status string) error {
	status = strings.ToLower(strings.TrimSpace(status))

	if status != AssignmentAccepted && status != AssignmentDeclined {
		return fmt.Errorf(
			"%w: response must be accepted or declined",
			ErrInvalidInput,
		)
	}

	return nil
}

func ValidateOrderStatus(status string) error {
	status = strings.ToLower(strings.TrimSpace(status))

	switch status {
	case OrderPickedUp, OrderInTransit, OrderDelivered:
		return nil
	default:
		return fmt.Errorf(
			"%w: courier can only set status to picked_up, in_transit, or delivered",
			ErrInvalidInput,
		)
	}
}

func ValidateStatusNotes(notes string) error {
	if utf8.RuneCountInString(notes) > MaxStatusNotesLength {
		return fmt.Errorf(
			"%w: status notes cannot exceed %d characters",
			ErrInvalidInput,
			MaxStatusNotesLength,
		)
	}

	return nil
}

func ValidateStatusTransition(oldStatus, newStatus string) error {
	oldStatus = strings.ToLower(strings.TrimSpace(oldStatus))
	newStatus = strings.ToLower(strings.TrimSpace(newStatus))

	switch oldStatus {
	case OrderAssigned:
		if newStatus == OrderPickedUp {
			return nil
		}
	case OrderPickedUp:
		if newStatus == OrderInTransit {
			return nil
		}
	case OrderInTransit:
		if newStatus == OrderDelivered {
			return nil
		}
	}

	return ErrInvalidTransition
}
