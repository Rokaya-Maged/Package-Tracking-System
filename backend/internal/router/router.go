package router

import (
    "net/http"

    "tracking-system/internal/orders"
)

func New(
    orderService *orders.OrderService,
	//
	//
) http.Handler {

    mux := http.NewServeMux()

    // Orders routes
    orderHandler := orders.New(orderService)
    mux.Handle("/api/orders", orderHandler)
    mux.Handle("/api/orders/", orderHandler)

    // Users routes
	// 

    // Health check
    mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Backend is running!"))
    })

    return mux
}