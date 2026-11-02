
package courier

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrInvalidInput = errors.New("invalid input")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// GetAssignedOrders returns the authenticated courier's
// pending and accepted assignments.
func (s *Service) GetAssignedOrders(
	ctx context.Context,
	courierID int64,
) ([]CourierAssignedOrder, error) {

	if courierID <= 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.GetAssignedOrders(ctx, courierID)
}

// GetOrderDetails returns an order assigned to the courier.
func (s *Service) GetOrderDetails(
	ctx context.Context,
	courierID int64,
	orderID int64,
) (*CourierAssignedOrder, error) {

	if courierID <= 0 || orderID <= 0 {
		return nil, ErrInvalidInput
	}

	return s.repo.GetOrderDetails(ctx, courierID, orderID)
}

// AcceptAssignment accepts a pending courier assignment.
func (s *Service) AcceptAssignment(
	ctx context.Context,
	courierID int64,
	orderID int64,
) error {

	if courierID <= 0 || orderID <= 0 {
		return ErrInvalidInput
	}

	return s.repo.RespondToAssignment(
		ctx,
		courierID,
		orderID,
		AssignmentAccepted,
	)
}

// DeclineAssignment declines a pending courier assignment.
func (s *Service) DeclineAssignment(
	ctx context.Context,
	courierID int64,
	orderID int64,
) error {

	if courierID <= 0 || orderID <= 0 {
		return ErrInvalidInput
	}

	return s.repo.RespondToAssignment(
		ctx,
		courierID,
		orderID,
		AssignmentDeclined,
	)
}

// UpdateOrderStatus updates the delivery progress.
//
// Allowed transitions:
// assigned  -> picked_up
// picked_up -> in_transit
// in_transit -> delivered
func (s *Service) UpdateOrderStatus(
	ctx context.Context,
	courierID int64,
	orderID int64,
	newStatus string,
	notes string,
) error {

	if courierID <= 0 || orderID <= 0 {
		return ErrInvalidInput
	}

	newStatus = strings.ToLower(strings.TrimSpace(newStatus))
	notes = strings.TrimSpace(notes)

	switch newStatus {
	case OrderPickedUp, OrderInTransit, OrderDelivered:
		// These are the only statuses a courier can set.
	default:
		return ErrInvalidInput
	}

	return s.repo.UpdateCourierOrderStatus(
		ctx,
		courierID,
		orderID,
		newStatus,
		notes,
	)
}

// isValidTransition checks the allowed shipment lifecycle.
func isValidTransition(oldStatus, newStatus string) bool {
	switch oldStatus {
	case OrderAssigned:
		return newStatus == OrderPickedUp

	case OrderPickedUp:
		return newStatus == OrderInTransit

	case OrderInTransit:
		return newStatus == OrderDelivered

	default:
		return false
	}
}
