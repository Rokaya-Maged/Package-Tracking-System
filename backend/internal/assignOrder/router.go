package assignOrder


import (
	"net/http"
)

func New(service *AssignmentService) http.Handler {

	handler := &AssignmentHandler{Service: service}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/assignments", handler.AssignOrderToCourierHandler)

	mux.HandleFunc("POST /api/assignments/reassign", handler.ReassignOrderToCourierHandler)

	mux.HandleFunc("GET /api/assignments", handler.GetAllAssignmetHandler)

	mux.HandleFunc("GET /api/assignments/order/{order_id}", handler.GetAssignmentsByOrderID)

	mux.HandleFunc("GET /api/assignments/courier/{courier_id}", handler.GetAssignmentsByCourierID)

	mux.HandleFunc("GET /api/assignments/status", handler.GetAssignmentsByStatus)

	return mux
}
