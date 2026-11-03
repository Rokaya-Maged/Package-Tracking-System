package assignOrder

import (
	"errors"
	"strings"
)

type AssignmentService struct {
	Repo *AssignmentRepository
}

// AssignOrderToCourier assigns an order to a courier.
func (s *AssignmentService) AssignOrderToCourier(orderID, courierID, assignedBy int) error {

	err := ValidateAssignment(orderID, courierID, assignedBy)
	if err != nil {
		return err
	}

	courier, err := s.Repo.GetUserRole(courierID)
	if err != nil {
		return err
	}
	if courier != "courier" {
		return ErrUserNotCourier
	}

	admin, err := s.Repo.GetUserRole(assignedBy)
	if err != nil {
		return err
	}
	if admin != "admin" {
		return ErrUserNotAdmin
	}

	return s.Repo.AssignOrderToCourier(orderID, courierID, assignedBy)
}

// ReassignOrderToCourier reassigns an order to a new courier.
func (s *AssignmentService) ReassignOrderToCourier(orderID, newCourierID, assignedBy int) error {

	err := ValidateAssignment(orderID, newCourierID, assignedBy)
	if err != nil {
		return err
	}

	courier, err := s.Repo.GetUserRole(newCourierID)
	if err != nil {
		return err
	}
	if courier != "courier" {
		return ErrUserNotCourier
	}

	admin, err := s.Repo.GetUserRole(assignedBy)
	if err != nil {
		return err
	}
	if admin != "admin" {
		return ErrUserNotAdmin
	}

	return s.Repo.ReassignOrderToCourier(orderID, newCourierID, assignedBy)
}

// get all assignments
func (s *AssignmentService) GetAllAssignments() ([]Assignment, error) {
	return s.Repo.GetAllAssignments()
}

// get assignments by order id
func (s *AssignmentService) GetAssignmentsByOrderID(orderID int) ([]Assignment, error) {
	if orderID <= 0 {
		return nil, errors.New("invalid order ID")
	}
	return s.Repo.GetAssignmentsByOrderID(orderID)
}

// get assignments by courier id
func (s *AssignmentService) GetAssignmentsByCourierID(courierID int) ([]Assignment, error) {
	if courierID <= 0 {
		return nil, errors.New("invalid courier ID")
	}

	courier, err := s.Repo.GetUserRole(courierID)
	if err != nil {
		return nil, err
	}
	if courier != "courier" {
		return nil, ErrUserNotCourier
	}

	return s.Repo.GetAssignmentsByCourierID(courierID)
}

func (s *AssignmentService) GetAssignmentsByStatus(status string) ([]Assignment, error) {
	validStatuses := map[string]bool{
		"pending":    true,
		"accepted":   true,
		"declined":   true,
		"reassigned": true,
	}

	status = strings.TrimSpace(status)

	if !validStatuses[status] {
		return nil, ErrInvalidAssignmentStatus
	}

	return s.Repo.GetAssignmentsByStatus(status)
}
