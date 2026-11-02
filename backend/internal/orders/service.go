package orders

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type OrderService struct {
	Repo *OrderRepository
}

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrInvalidOrderData   = errors.New("invalid order data")
	ErrInvalidOrderStatus = errors.New("invalid order status")
)

// get all order
func (s *OrderService) GetAllOrders() ([]Order, error) {
	return s.Repo.GetAllOrders()
}

// get order by id
func (s *OrderService) GetOrderByID(id int) (*Order, error) {

	if id <= 0 {
		return nil, errors.New("invalid order ID")
	}

	order, err := s.Repo.GetOrderByID(id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	return order, nil
}

// create order
func (s *OrderService) AddOrder(order *Order) error {
	if order == nil {
		return fmt.Errorf("%w: order cannot be nil", ErrInvalidOrderData)
	}

	if order.CustomerID <= 0 {
		return fmt.Errorf("%w: invalid customer ID", ErrInvalidOrderData)
	}

	if strings.TrimSpace(order.DeliveryAddress) == "" {
		return fmt.Errorf("%w: delivery address is required", ErrInvalidOrderData)
	}

	if strings.TrimSpace(order.PickupAddress) == "" {
		return fmt.Errorf("%w: pickup address is required", ErrInvalidOrderData)
	}

	return s.Repo.CreateOrder(order)
}

// update order status
func (s *OrderService) UpdateOrderStatus(id int, status string) error {
	if id <= 0 {
		return errors.New("invalid order ID")
	}

	validStatuses := map[string]bool{
		"Picked Up":  true,
		"In Transit": true,
		"Delivered":  true,
	}

	if !validStatuses[status] {
		return ErrInvalidOrderStatus
	}

	err := s.Repo.UpdateOrderStatus(id, status)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrOrderNotFound
	}

	return err
}

// Delete order
func (s *OrderService) DeleteOrder(id int) error {

	if id <= 0 {
		return errors.New("invalid order ID")
	}

	err := s.Repo.DeleteOrder(id)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrOrderNotFound
	}

	return err
}
