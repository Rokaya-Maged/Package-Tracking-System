
package courier

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound          = errors.New("assignment or order not found")
	ErrInvalidAssignment = errors.New("assignment is not pending")
	ErrInvalidTransition = errors.New("invalid order status transition")
	ErrNotAccepted       = errors.New("assignment has not been accepted")
)

type Repository struct {
	DB *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{DB: db}
}

// GetAssignedOrders returns pending and accepted assignments
// belonging to the authenticated courier.
func (r *Repository) GetAssignedOrders(
	ctx context.Context,
	courierID int64,
) ([]CourierAssignedOrder, error) {

	query := `
		SELECT
			o.id,
			o.tracking_number,
			o.customer_name,
			o.customer_phone,
			o.pickup_address,
			o.dropoff_address,
			o.delivery_notes,
			o.package_description,
			o.package_weight_kg,
			o.payment_type,
			o.cod_amount,
			o.delivery_fee,
			o.preferred_delivery_at,
			o.status,
			ca.id,
			ca.status,
			ca.assigned_at
		FROM courier_assignments ca
		INNER JOIN orders o ON o.id = ca.order_id
		WHERE ca.courier_id = $1
		  AND ca.status IN ('pending', 'accepted')
		ORDER BY ca.assigned_at DESC
	`

	rows, err := r.DB.Query(ctx, query, courierID)
	if err != nil {
		return nil, fmt.Errorf("get courier assignments: %w", err)
	}
	defer rows.Close()

	orders := make([]CourierAssignedOrder, 0)

	for rows.Next() {
		var order CourierAssignedOrder

		err := rows.Scan(
			&order.OrderID,
			&order.TrackingNumber,
			&order.CustomerName,
			&order.CustomerPhone,
			&order.PickupAddress,
			&order.DropoffAddress,
			&order.DeliveryNotes,
			&order.PackageDescription,
			&order.PackageWeightKG,
			&order.PaymentType,
			&order.CODAmount,
			&order.DeliveryFee,
			&order.PreferredDeliveryAt,
			&order.OrderStatus,
			&order.AssignmentID,
			&order.AssignmentStatus,
			&order.AssignedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("scan courier assignment: %w", err)
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate courier assignments: %w", err)
	}

	return orders, nil
}

// GetOrderDetails returns shipment information only if the
// order is assigned to this courier and the assignment is active.
func (r *Repository) GetOrderDetails(
	ctx context.Context,
	courierID int64,
	orderID int64,
) (*CourierAssignedOrder, error) {

	query := `
		SELECT
			o.id,
			o.tracking_number,
			o.customer_name,
			o.customer_phone,
			o.pickup_address,
			o.dropoff_address,
			o.delivery_notes,
			o.package_description,
			o.package_weight_kg,
			o.payment_type,
			o.cod_amount,
			o.delivery_fee,
			o.preferred_delivery_at,
			o.status,
			ca.id,
			ca.status,
			ca.assigned_at
		FROM courier_assignments ca
		INNER JOIN orders o ON o.id = ca.order_id
		WHERE ca.courier_id = $1
		  AND o.id = $2
		  AND ca.status IN ('pending', 'accepted')
	`

	var order CourierAssignedOrder

	err := r.DB.QueryRow(
		ctx,
		query,
		courierID,
		orderID,
	).Scan(
		&order.OrderID,
		&order.TrackingNumber,
		&order.CustomerName,
		&order.CustomerPhone,
		&order.PickupAddress,
		&order.DropoffAddress,
		&order.DeliveryNotes,
		&order.PackageDescription,
		&order.PackageWeightKG,
		&order.PaymentType,
		&order.CODAmount,
		&order.DeliveryFee,
		&order.PreferredDeliveryAt,
		&order.OrderStatus,
		&order.AssignmentID,
		&order.AssignmentStatus,
		&order.AssignedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get courier order details: %w", err)
	}

	return &order, nil
}

// RespondToAssignment accepts or declines a pending assignment.
//
// The assignment row is locked to prevent concurrent responses.
func (r *Repository) RespondToAssignment(
	ctx context.Context,
	courierID int64,
	orderID int64,
	newStatus string,
) error {

	if newStatus != AssignmentAccepted &&
		newStatus != AssignmentDeclined {
		return ErrInvalidAssignment
	}

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin assignment transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var assignmentID int64
	var currentStatus string

	err = tx.QueryRow(ctx, `
		SELECT id, status
		FROM courier_assignments
		WHERE order_id = $1
		  AND courier_id = $2
		FOR UPDATE
	`, orderID, courierID).Scan(
		&assignmentID,
		&currentStatus,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	if err != nil {
		return fmt.Errorf("find courier assignment: %w", err)
	}

	if currentStatus != AssignmentPending {
		return ErrInvalidAssignment
	}

	// Only the assignment status is changed here.
	// The order status remains "assigned".
	_, err = tx.Exec(ctx, `
		UPDATE courier_assignments
		SET status = $1,
		    responded_at = NOW()
		WHERE id = $2
	`, newStatus, assignmentID)

	if err != nil {
		return fmt.Errorf("update assignment response: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit assignment response: %w", err)
	}

	return nil
}

// UpdateCourierOrderStatus updates an order only when the
// courier has an accepted assignment.
//
// Updating the order and inserting its history record are
// performed in the same transaction.
func (r *Repository) UpdateCourierOrderStatus(
	ctx context.Context,
	courierID int64,
	orderID int64,
	newStatus string,
	notes string,
) error {

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin status transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var oldStatus string

	// Lock both the order and its accepted assignment.
	err = tx.QueryRow(ctx, `
		SELECT o.status
		FROM orders o
		INNER JOIN courier_assignments ca
			ON ca.order_id = o.id
		WHERE o.id = $1
		  AND ca.courier_id = $2
		  AND ca.status = 'accepted'
		FOR UPDATE OF o, ca
	`, orderID, courierID).Scan(&oldStatus)

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	if err != nil {
		return fmt.Errorf("verify courier assignment: %w", err)
	}

	if !isValidTransition(oldStatus, newStatus) {
		return ErrInvalidTransition
	}

	// Update the shipment status.
	result, err := tx.Exec(ctx, `
		UPDATE orders
		SET status = $1,
		    updated_at = NOW()
		WHERE id = $2
		  AND status = $3
	`, newStatus, orderID, oldStatus)

	if err != nil {
		return fmt.Errorf("update shipment status: %w", err)
	}

	if result.RowsAffected() != 1 {
		return ErrInvalidTransition
	}

	// Insert the status history record.
	_, err = tx.Exec(ctx, `
		INSERT INTO order_status_history (
			order_id,
			old_status,
			new_status,
			changed_by,
			notes
		)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''))
	`, orderID, oldStatus, newStatus, courierID, notes)

	if err != nil {
		return fmt.Errorf("insert order status history: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit shipment status transaction: %w", err)
	}

	return nil
}
