package orders

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	DB *pgxpool.Pool
}

// get all order
func (r *OrderRepository) GetAllOrders() ([]Order, error) {
	query := `
		SELECT id, customer_id, status, total_price, delivery_address, pickup_address, created_at
		FROM orders
		ORDER BY id
	`
	rows, err := r.DB.Query(context.Background(), query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var orders []Order

	for rows.Next() {
		var order Order

		err := rows.Scan(
			&order.ID,
			&order.CustomerID,
			&order.Status,
			// &order.TotalPrice,
			&order.DeliveryAddress,
			&order.PickupAddress,
			&order.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

// get order by id
func (r *OrderRepository) GetOrderByID(id int) (*Order, error) {
	query := `
		SELECT id, customer_id, status, total_price, delivery_address, pickup_address, created_at
		FROM orders
		WHERE orders.id = $1
	`

	var order Order

	err := r.DB.QueryRow(context.Background(), query, id).Scan(
		&order.ID,
		&order.CustomerID,
		&order.Status,
		// &order.TotalPrice,
		&order.DeliveryAddress,
		&order.PickupAddress,
		&order.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &order, nil
}

// create new order
func (r *OrderRepository) CreateOrder(order *Order) error {
	query := `
		INSERT INTO orders (customer_id, status, total_price, delivery_address, pickup_address)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	return r.DB.QueryRow(context.Background(), query,
		order.CustomerID,
		order.Status,
		// order.TotalPrice,
		order.DeliveryAddress,
		order.PickupAddress,
	).Scan(&order.ID, &order.CreatedAt)

}

// Update order status
func (r *OrderRepository) UpdateOrderStatus(id int, status string) error {
	query := `
		UPDATE orders
		SET status = $1
		WHERE id = $2
	`

	result, err := r.DB.Exec(context.Background(), query, status, id)
	if err != nil {
		return err
	}

	// retrieve the number of rows affected by the update operation
	rowsAffected := result.RowsAffected()

	// check if any rows were affected, if not return sql.ErrNoRows
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// delete order
func (r *OrderRepository) DeleteOrder(id int) error {
	query := `
		DELETE FROM orders
		WHERE id = $1
	`
	result, err := r.DB.Exec(context.Background(), query, id)

	rowsAffected := result.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
