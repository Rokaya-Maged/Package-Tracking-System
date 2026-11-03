
package courier

import (
	"context"
	"strings"
)

type Service struct {
	Repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{Repo: repo}
}

// GetAssignedOrders returns all active assignments for a courier.
func (s *Service) GetAssignedOrders(
	ctx context.Context,
	courierID int64,
) ([]CourierAssignedOrder, error) {

	if err := ValidateCourierID(courierID); err != nil {
		return nil, err
	}

	return s.Repo.GetAssignedOrders(ctx, courierID)
}

// GetOrderDetails returns the details of one order assigned to the courier.
func (s *Service) GetOrderDetails(
	ctx context.Context,
	courierID int64,
	orderID int64,
) (*CourierAssignedOrder, error) {

	if err := ValidateCourierID(courierID); err != nil {
		return nil, err
	}

	if err := ValidateOrderID(orderID); err != nil {
		return nil, err
	}

	return s.Repo.GetOrderDetails(ctx, courierID, orderID)
}

// AcceptAssignment accepts a pending assignment.
func (s *Service) AcceptAssignment(
	ctx context.Context,
	courierID int64,
	orderID int64,
) error {

	if err := ValidateCourierID(courierID); err != nil {
		return err
	}

	if err := ValidateOrderID(orderID); err != nil {
		return err
	}

	return s.Repo.RespondToAssignment(
		ctx,
		courierID,
		orderID,
		AssignmentAccepted,
	)
}

// DeclineAssignment declines a pending assignment.
func (s *Service) DeclineAssignment(
	ctx context.Context,
	courierID int64,
	orderID int64,
) error {

	if err := ValidateCourierID(courierID); err != nil {
		return err
	}

	if err := ValidateOrderID(orderID); err != nil {
		return err
	}

	return s.Repo.RespondToAssignment(
		ctx,
		courierID,
		orderID,
		AssignmentDeclined,
	)
}

// UpdateOrderStatus updates the delivery lifecycle status.
func (s *Service) UpdateOrderStatus(
	ctx context.Context,
	courierID int64,
	orderID int64,
	status string,
	notes string,
) error {

	if err := ValidateCourierID(courierID); err != nil {
		return err
	}

	if err := ValidateOrderID(orderID); err != nil {
		return err
	}

	status = strings.ToLower(strings.TrimSpace(status))
	notes = strings.TrimSpace(notes)

	if err := ValidateOrderStatus(status); err != nil {
		return err
	}

	if err := ValidateStatusNotes(notes); err != nil {
		return err
	}

	return s.Repo.UpdateCourierOrderStatus(
		ctx,
		courierID,
		orderID,
		status,
		notes,
	)
}