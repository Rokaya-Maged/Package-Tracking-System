package assignOrder

import "errors"

var (
	ErrInvalidOrderID    = errors.New("order ID is required and must be positive")
	ErrInvalidCourierID  = errors.New("courier ID is required and must be positive")
	ErrInvalidAssignedBy = errors.New("assignedBy is required and must be positive")

	ErrUserNotCourier = errors.New("user is not a courier")
	ErrUserNotAdmin   = errors.New("user is not an admin")
	ErrUserNotFound   = errors.New("user not found")

	ErrOrderNotFound        = errors.New("order not found")
	ErrOrderNotPending      = errors.New("order is not in pending status")
	ErrAssignmentFailed     = errors.New("failed to assign order")
	ErrOrderAlreadyAssigned = errors.New("order is already assigned")
	ErrNoActiveAssignment   = errors.New("no active assignment found")
	ErrSameCourier          = errors.New("order is already assigned to this courier")
    ErrInvalidAssignmentStatus = errors.New("invalid assignment status")
)

func ValidateAssignment(orderID, courierID, assignedBy int) error {

	if orderID <= 0 {
		return ErrInvalidOrderID
	}

	if courierID <= 0 {
		return ErrInvalidCourierID
	}

	if assignedBy <= 0 {
		return ErrInvalidAssignedBy
	}

	return nil
}
