package assignOrder

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AssignmentRepository struct {
	DB *pgxpool.Pool
}

func (r *AssignmentRepository) GetUserRole(userID int) (string, error) {

	query := `
        SELECT role
        FROM users
        WHERE id = $1
    `

	var role string

	err := r.DB.QueryRow(context.Background(), query, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("%w: user with ID %d", ErrUserNotFound, userID)
		}

		return "", fmt.Errorf("failed to retrieve user role: %w", err)
	}

	return role, nil
}

func (r *AssignmentRepository) AssignOrderToCourier(orderID int, courierID int, assignedBy int) error {

	// create context for database operations
	ctx := context.Background()

	// start a transaction to execute both operations together
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer tx.Rollback(ctx) // rollback in case of error

	query := `
        INSERT INTO courier_assignments
        (order_id, courier_id, assigned_by, status, assigned_at)
        VALUES ($1, $2, $3, 'pending', NOW())
    `

	_, err = tx.Exec(ctx, query, orderID, courierID, assignedBy)

	if err != nil {

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {

			switch pgErr.Code {

			case "23505":
				// Unique violation
				return ErrOrderAlreadyAssigned

			case "23503":
				// Foreign key violation
				return fmt.Errorf("invalid order, courier, or assigned user ID: %w", err)

			case "23514":
				// Check constraint violation
				return fmt.Errorf("invalid assignment data: %w", err)

			case "23502":
				// Not null violation
				return fmt.Errorf("required assignment field is missing: %w", err)
			}
		}

		return fmt.Errorf("%w: %w", ErrAssignmentFailed, err)
	}

	// Update order status
	updateOrderQuery := `
		UPDATE orders
        SET status = 'assigned', updated_at = NOW()
        WHERE id = $1 AND status = 'pending'
	`

	result, err := tx.Exec(ctx, updateOrderQuery, orderID)

	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrOrderNotFound
	}

	// commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ReassignOrderToCourier reassigns an order to a new courier
func (r *AssignmentRepository) ReassignOrderToCourier(OrderID int, newCourierID int, reassignedBy int) error {

	// create context for database operations
	ctx := context.Background()

	// start a transaction to execute both operations together
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer tx.Rollback(ctx) // rollback in case of error

	// retrieve the current assignment and courier ID
	var currentCourierID int

	// select the current courier ID to check if the order is already assigned to the new courier
	selectQuery := `
		SELECT ca.courier_id
		FROM courier_assignments ca JOIN orders o ON o.id = ca.order_id
		WHERE ca.order_id = $1 AND ca.status IN ('pending', 'accepted') AND o.status IN ('assigned')
		FOR UPDATE OF ca, o
	`

	err = tx.QueryRow(ctx, selectQuery, OrderID).Scan(&currentCourierID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoActiveAssignment
		}
		return fmt.Errorf("failed to retrieve current assignment: %w", err)
	}

	// check if the new courier is the same as the current courier
	if currentCourierID == newCourierID {
		return ErrSameCourier
	}

	// update the existing assignment to mark it as reassigned
	updateQuery := `
		UPDATE courier_assignments
		SET status = 'reassigned'
		WHERE order_id = $1 AND status IN ('pending', 'accepted')
	`

	result, err := tx.Exec(ctx, updateQuery, OrderID)
	if err != nil {
		return fmt.Errorf("failed to update existing assignment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNoActiveAssignment
	}

	// insert a new assignment for the new courier
	insertQuery := `
		INSERT INTO courier_assignments (order_id, courier_id, assigned_by, status, assigned_at)
		VALUES ($1, $2, $3, 'pending', NOW())
	`

	_, err = tx.Exec(ctx, insertQuery, OrderID, newCourierID, reassignedBy)
	if err != nil {
		return fmt.Errorf("failed to create new assignment: %w", err)
	}

	// commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// get all assignments
func (r *AssignmentRepository) GetAllAssignments() ([]Assignment, error) {

	query := `
		SELECT id, order_id, courier_id, assigned_by, status, assigned_at
		FROM courier_assignments
	`
	rows, err := r.DB.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve assignments: %w", err)
	}

	defer rows.Close()

	var assignments []Assignment

	for rows.Next() {
		var assignment Assignment
		err := rows.Scan(&assignment.ID, &assignment.OrderID, &assignment.CourierID, &assignment.AssignedBy, &assignment.Status, &assignment.AssignedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}
		assignments = append(assignments, assignment)

	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read assignments: %w", err)
	}

	return assignments, nil
}

// get assignments by order id
func (r *AssignmentRepository) GetAssignmentsByOrderID(orderID int) ([]Assignment, error) {

	query := `
		SELECT id, order_id, courier_id, assigned_by, status, assigned_at
		FROM courier_assignments
		WHERE order_id = $1
	`
	rows, err := r.DB.Query(context.Background(), query, orderID)

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve assignments for order ID %d: %w", orderID, err)
	}
	defer rows.Close()

	// the same order can be assigned to multiple couriers(RE-ASSIGNMENT), so we need to return a slice of assignments
	var assignments []Assignment

	for rows.Next() {
		var assignment Assignment
		err := rows.Scan(&assignment.ID, &assignment.OrderID, &assignment.CourierID, &assignment.AssignedBy, &assignment.Status, &assignment.AssignedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}
		assignments = append(assignments, assignment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read assignments: %w", err)
	}

	return assignments, nil
}

// get assignments by courier id
func (r *AssignmentRepository) GetAssignmentsByCourierID(courierID int) ([]Assignment, error) {

	query := `
		SELECT id, order_id, courier_id, assigned_by, status, assigned_at
		FROM courier_assignments
		WHERE courier_id = $1
	`
	rows, err := r.DB.Query(context.Background(), query, courierID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve assignments for courier ID %d: %w", courierID, err)
	}

	defer rows.Close()

	var assignments []Assignment

	for rows.Next() {
		var assignment Assignment
		err := rows.Scan(&assignment.ID, &assignment.OrderID, &assignment.CourierID, &assignment.AssignedBy, &assignment.Status, &assignment.AssignedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}
		assignments = append(assignments, assignment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read assignments: %w", err)
	}

	return assignments, nil
}

// Get assignments by status
func (r *AssignmentRepository) GetAssignmentsByStatus(status string) ([]Assignment, error) {

	query := `
        SELECT id, order_id, courier_id, assigned_by, status, assigned_at
        FROM courier_assignments
        WHERE status = $1
        ORDER BY assigned_at DESC
    `

	rows, err := r.DB.Query(context.Background(), query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve assignments by status: %w", err)
	}
	defer rows.Close()

	var assignments []Assignment

	for rows.Next() {
		var assignment Assignment

		err := rows.Scan(
			&assignment.ID,
			&assignment.OrderID,
			&assignment.CourierID,
			&assignment.AssignedBy,
			&assignment.Status,
			&assignment.AssignedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}

		assignments = append(assignments, assignment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read assignments: %w", err)
	}

	return assignments, nil
}
