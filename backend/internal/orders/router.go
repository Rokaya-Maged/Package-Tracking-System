package orders

import (
	"net/http"
)

func New(service *OrderService) http.Handler {

	handler := &OrderHandler{Service: service}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/orders", handler.GetAllOrdersHandler)

	mux.HandleFunc("GET /api/orders/{id}", handler.GetOrderByIDHandler)

	mux.HandleFunc("POST /api/orders", handler.CreateOrderHandler)

	mux.HandleFunc("PUT /api/orders/{id}", handler.UpdateOrderHandler)

	mux.HandleFunc("DELETE /api/orders/{id}", handler.DeleteOrderHandler)

	return mux
}
