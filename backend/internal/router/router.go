package router

import (
	"errors"
	"net/http"

	"tracking-system/internal/assignOrder"
	"tracking-system/internal/auth"
	"tracking-system/internal/courier"
	"tracking-system/internal/middleware"
	"tracking-system/internal/orders"
)

func New(
	orderService *orders.OrderService,
	assignmentService *assignOrder.AssignmentService,
	courierService *courier.Service,
) http.Handler {

	mux := http.NewServeMux()
	// Registration route
	mux.HandleFunc("POST /api/auth/register", auth.Register)
	//login route
	mux.HandleFunc("POST /api/auth/login", auth.Login)

	// Orders routes
	orderHandler := orders.New(orderService)
	mux.Handle("/api/orders", orderHandler)
	mux.Handle("/api/orders/", orderHandler)
	mux.Handle("POST /api/orders", middleware.RequireAuth(middleware.RequireRole("business", orderHandler)))
	mux.Handle("GET /api/orders", middleware.RequireAuth(orders.OrdersListByRole(orderService, orderHandler)))
	mux.Handle("PUT /api/orders/{id}", middleware.RequireAuth(middleware.RequireRole("admin", orderHandler)))
	mux.Handle("DELETE /api/orders/{id}", middleware.RequireAuth(middleware.RequireRole("admin", orderHandler)))
	mux.Handle("GET /api/orders/{id}", middleware.RequireAuth(orders.OrderDetailsByRole(orderService, orderHandler)))

	// Assignments routes
	assignmentHandler := assignOrder.New(assignmentService)
	mux.Handle("/api/assignments", middleware.RequireAuth(middleware.RequireRole("admin", assignmentHandler)))
	mux.Handle("/api/assignments/", middleware.RequireAuth(middleware.RequireRole("admin", assignmentHandler)))
	mux.Handle("GET /api/couriers", middleware.RequireAuth(middleware.RequireRole("admin", assignmentHandler)))

	// Courier routes
	courierHandler := courier.New(courierService, func(r *http.Request) (int64, error) {
		user, ok := middleware.GetUser(r)
		if !ok {
			return 0, errors.New("authenticated user not found")
		}
		return user.ID, nil
	})
	mux.Handle(
		"/api/courier/",
		middleware.RequireAuth(middleware.RequireRole("courier", courierHandler)),
	)

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Backend is running!"))
	})

	return mux
}
