
package router

import (
	"net/http"

	"tracking-system/internal/assignOrder"
	"tracking-system/internal/courier"
	"tracking-system/internal/orders"
)

func New(
	orderService *orders.OrderService,
	assignmentService *assignOrder.AssignmentService,
	courierService *courier.Service,
	resolveUserID courier.UserIDResolver,
) http.Handler {

	mux := http.NewServeMux()

	// Orders routes
	orderHandler := orders.New(orderService)
	mux.Handle("/api/orders", orderHandler)
	mux.Handle("/api/orders/", orderHandler)

	// Assignments routes
	assignmentHandler := assignOrder.New(assignmentService)
	mux.Handle("/api/assignments", assignmentHandler)
	mux.Handle("/api/assignments/", assignmentHandler)

	// Courier routes
	courierHandler := courier.New(courierService, resolveUserID)
	mux.Handle("/api/courier/", courierHandler)

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Backend is running!"))
	})

	return mux
}
