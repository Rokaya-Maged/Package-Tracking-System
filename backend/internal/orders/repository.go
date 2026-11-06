package orders

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	DB *pgxpool.Pool
}

type orderScanner interface {
	Scan(dest ...any) error
}

// get all order
func (r *OrderRepository) GetAllOrders() ([]Order, error) {
	query := `
        SELECT id, tracking_number, business_user_id, customer_name, customer_phone,
               pickup_address, dropoff_address, delivery_notes, package_description,
               package_weight_kg, preferred_delivery_at, item_price, delivery_fee,
               delivery_fee_payer, payment_type, total_customer_price, cod_amount,
               cod_collected_at, status, created_at, updated_at
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
			&order.TrackingNumber,
			&order.BusinessUserID,
			&order.CustomerName,
			&order.CustomerPhone,
			&order.PickupAddress,
			&order.DropoffAddress,
			&order.DeliveryNotes,
			&order.PackageDescription,
			&order.PackageWeightKg,
			&order.PreferredDeliveryAt,
			&order.ItemPrice,
			&order.DeliveryFee,
			&order.DeliveryFeePayer,
			&order.PaymentType,
			&order.TotalCustomerPrice,
			&order.CODAmount,
			&order.CODCollectedAt,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
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
// get order by id with courier information
func (r *OrderRepository) GetOrderByID(id int) (*Order, error) {
	query := `
		SELECT o.id, o.tracking_number, o.business_user_id,
		       o.customer_name, o.customer_phone,
		       o.pickup_address, o.dropoff_address,
		       o.delivery_notes, o.package_description,
		       o.package_weight_kg, o.preferred_delivery_at,
		       o.item_price, o.delivery_fee,
		       o.delivery_fee_payer, o.payment_type,
		       o.total_customer_price, o.cod_amount,
		       o.cod_collected_at, o.status,
		       o.created_at, o.updated_at,
		       courier.full_name, courier.phone
		FROM orders o
		LEFT JOIN LATERAL (
			SELECT u.full_name, u.phone
			FROM courier_assignments ca
			JOIN users u ON u.id = ca.courier_id
			WHERE ca.order_id = o.id
			  AND ca.status IN ('pending', 'accepted')
			ORDER BY ca.assigned_at DESC, ca.id DESC
			LIMIT 1
		) courier ON TRUE
		WHERE o.id = $1
	`

	var order Order

	err := r.DB.QueryRow(context.Background(), query, id).Scan(
		&order.ID,
		&order.TrackingNumber,
		&order.BusinessUserID,
		&order.CustomerName,
		&order.CustomerPhone,
		&order.PickupAddress,
		&order.DropoffAddress,
		&order.DeliveryNotes,
		&order.PackageDescription,
		&order.PackageWeightKg,
		&order.PreferredDeliveryAt,
		&order.ItemPrice,
		&order.DeliveryFee,
		&order.DeliveryFeePayer,
		&order.PaymentType,
		&order.TotalCustomerPrice,
		&order.CODAmount,
		&order.CODCollectedAt,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.CourierName,
		&order.CourierPhone,
	)

	if err != nil {
		return nil, err
	}

	return &order, nil
}

// create new order
func (r *OrderRepository) CreateOrder(order *Order) error {
	query := `
        INSERT INTO orders (
            tracking_number, business_user_id, customer_name, customer_phone,
            pickup_address, dropoff_address, delivery_notes, package_description,
            package_weight_kg, preferred_delivery_at, item_price, delivery_fee,
            delivery_fee_payer, payment_type
        )
        VALUES (
            $1, $2, $3, $4, $5, $6, $7,
            $8, $9, $10, $11, $12, $13, $14
        )
        RETURNING id, total_customer_price, cod_amount, status, created_at, updated_at
    `

	err := r.DB.QueryRow(context.Background(), query,
		order.TrackingNumber,
		order.BusinessUserID,
		order.CustomerName,
		order.CustomerPhone,
		order.PickupAddress,
		order.DropoffAddress,
		order.DeliveryNotes,
		order.PackageDescription,
		order.PackageWeightKg,
		order.PreferredDeliveryAt,
		order.ItemPrice,
		order.DeliveryFee,
		order.DeliveryFeePayer,
		order.PaymentType,
	).Scan(
		&order.ID,
		&order.TotalCustomerPrice,
		&order.CODAmount,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" && pgErr.ConstraintName == "orders_business_user_id_fkey" {
				return ErrBusinessUserNotFound
			}

			if pgErr.Code == "23505" && pgErr.ConstraintName == "orders_tracking_number_key" {
				return ErrTrackingNumberExists
			}
		}
		return err
	}

	return nil
}

// Update order status
func (r *OrderRepository) UpdateOrderStatus(id int, status string) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = NOW()
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
func (r *OrderRepository) CancelPendingOrder(id int, businessUserID int64) (bool, error) {
	query := `
		UPDATE orders
		SET status = 'cancelled', updated_at = NOW()
		WHERE id = $1
		  AND business_user_id = $2
		  AND status = 'pending'
	`

	result, err := r.DB.Exec(context.Background(), query, id, businessUserID)

	if err != nil {
		return false, err
	}

	return result.RowsAffected() == 1, nil
}
