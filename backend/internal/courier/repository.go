
package courier

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	DB *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{DB: db}
}

// GetAssignedOrders returns orders assigned to the specified courier.
// Pending assignments can be accepted or declined.
// Accepted assignments can proceed through delivery.
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
			o.preferred_delivery_at,
			o.item_price,
			o.delivery_fee,
			o.delivery_fee_payer,
			o.payment_type,
			o.total_customer_price,
			o.cod_amount,
			o.cod_collected_at,
			o.status,
			ca.id,
			ca.status,
			ca.assigned_at
		FROM courier_assignments ca
		INNER JOIN orders o ON o.id = ca.order_id
		WHERE ca.courier_id = $1
		  AND ca.status IN ($2, $3)
		ORDER BY ca.assigned_at DESC
	`

	rows, err := r.DB.Query(
		ctx,
		query,
		courierID,
		AssignmentPending,
		AssignmentAccepted,
	)
	if err != nil {
		return nil, fmt.Errorf("get assigned orders: %w", err)
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
			&order.PreferredDeliveryAt,
			&order.ItemPrice,
			&order.DeliveryFee,
			&order.DeliveryFeePayer,
			&order.PaymentType,
			&order.TotalCustomerPrice,
			&order.CODAmount,
			&order.CODCollectedAt,
			&order.OrderStatus,
			&order.AssignmentID,
			&order.AssignmentStatus,
			&order.AssignedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan assigned order: %w", err)
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate assigned orders: %w", err)
	}

	return orders, nil
}

// GetOrderDetails returns a specific order only if it is assigned
// to the authenticated courier and its assignment is still active.
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
			o.preferred_delivery_at,
			o.item_price,
			o.delivery_fee,
			o.delivery_fee_payer,
			o.payment_type,
			o.total_customer_price,
			o.cod_amount,
			o.cod_collected_at,
			o.status,
			ca.id,
			ca.status,
			ca.assigned_at
		FROM courier_assignments ca
		INNER JOIN orders o ON o.id = ca.order_id
		WHERE ca.courier_id = $1
		  AND ca.order_id = $2
		  AND ca.status IN ($3, $4)
	`

	var order CourierAssignedOrder

	err := r.DB.QueryRow(
		ctx,
		query,
		courierID,
		orderID,
		AssignmentPending,
		AssignmentAccepted,
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
		&order.PreferredDeliveryAt,
		&order.ItemPrice,
		&order.DeliveryFee,
		&order.DeliveryFeePayer,
		&order.PaymentType,
		&order.TotalCustomerPrice,
		&order.CODAmount,
		&order.CODCollectedAt,
		&order.OrderStatus,
		&order.AssignmentID,
		&order.AssignmentStatus,
		&order.AssignedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get order details: %w", err)
	}

	return &order, nil
}

// RespondToAssignment accepts or declines a pending assignment.
//
// The assignment and order are locked to prevent concurrent responses
// and to prevent accepting an assignment for a cancelled/completed order.
func (r *Repository) RespondToAssignment(
	ctx context.Context,
	courierID int64,
	orderID int64,
	newStatus string,
) error {

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin assignment response transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var assignmentID int64
	var currentAssignmentStatus string
	var orderStatus string

	query := `
		SELECT ca.id, ca.status, o.status
		FROM courier_assignments ca
		INNER JOIN orders o ON o.id = ca.order_id
		WHERE ca.order_id = $1
		  AND ca.courier_id = $2
		FOR UPDATE OF ca, o
	`

	err = tx.QueryRow(ctx, query, orderID, courierID).Scan(
		&assignmentID,
		&currentAssignmentStatus,
		&orderStatus,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("lock assignment: %w", err)
	}

	if currentAssignmentStatus != AssignmentPending {
		return ErrInvalidAssignment
	}

	if orderStatus != OrderAssigned {
		return ErrOrderNotAssignable
	}

	updateQuery := `
		UPDATE courier_assignments
		SET status = $1,
		    responded_at = NOW()
		WHERE id = $2
	`

	_, err = tx.Exec(ctx, updateQuery, newStatus, assignmentID)
	if err != nil {
		return fmt.Errorf("respond to assignment: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit assignment response: %w", err)
	}

	return nil
}

// UpdateCourierOrderStatus updates an order's delivery status
// and records the change in order_status_history.
//
// Only the courier with an accepted assignment can update the order.
func (r *Repository) UpdateCourierOrderStatus(
	ctx context.Context,
	courierID int64,
	orderID int64,
	newStatus string,
	notes string,
) error {

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin status update transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var oldStatus string

	query := `
		SELECT o.status
		FROM orders o
		INNER JOIN courier_assignments ca ON ca.order_id = o.id
		WHERE o.id = $1
		  AND ca.courier_id = $2
		  AND ca.status = $3
		FOR UPDATE OF o, ca
	`

	err = tx.QueryRow(
		ctx,
		query,
		orderID,
		courierID,
		AssignmentAccepted,
	).Scan(&oldStatus)

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("get order status: %w", err)
	}

	if err := ValidateStatusTransition(oldStatus, newStatus); err != nil {
		return err
	}

	updateQuery := `
		UPDATE orders
		SET status = $1,
		    updated_at = NOW()
		WHERE id = $2
	`

	_, err = tx.Exec(ctx, updateQuery, newStatus, orderID)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}

	historyQuery := `
		INSERT INTO order_status_history (
			order_id,
			old_status,
			new_status,
			changed_by,
			notes
		)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''))
	`

	_, err = tx.Exec(
		ctx,
		historyQuery,
		orderID,
		oldStatus,
		newStatus,
		courierID,
		notes,
	)
	if err != nil {
		return fmt.Errorf("insert order status history: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit order status update: %w", err)
	}

	return nil
}
