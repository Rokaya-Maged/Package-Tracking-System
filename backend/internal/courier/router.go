
package courier

import "net/http"

func New(service *Service, resolveUserID UserIDResolver) http.Handler {

	handler := NewHandler(service, resolveUserID)

	mux := http.NewServeMux()

	// Get all orders assigned to the logged-in courier.
	mux.HandleFunc("GET /api/courier/orders", handler.GetAssignedOrders)

	// Get details of a specific assigned order.
	mux.HandleFunc("GET /api/courier/orders/{id}", handler.GetOrderDetails)

	// Accept an assignment.
	mux.HandleFunc("PATCH /api/courier/orders/{id}/accept", handler.AcceptAssignment)

	// Decline an assignment.
	mux.HandleFunc("PATCH /api/courier/orders/{id}/decline", handler.DeclineAssignment)

	// Update delivery status.
	mux.HandleFunc("PATCH /api/courier/orders/{id}/status", handler.UpdateOrderStatus)

	return mux
}
