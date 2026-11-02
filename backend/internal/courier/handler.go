
package courier

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
)

type UserIDResolver func(r *http.Request) (int64, error)

type Handler struct {
	service   *Service
	getUserID UserIDResolver
}

func NewHandler(
	service *Service,
	resolver UserIDResolver,
) *Handler {
	return &Handler{
		service:   service,
		getUserID: resolver,
	}
}

// RegisterRoutes registers courier-specific endpoints.
//
// IMPORTANT:
// Apply authentication and courier-role middleware
// to these routes in the application's main router.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/courier/orders", h.GetAssignedOrders)

	mux.HandleFunc(
		"GET /api/courier/orders/{id}",
		h.GetOrderDetails,
	)

	mux.HandleFunc(
		"PATCH /api/courier/orders/{id}/accept",
		h.AcceptAssignment,
	)

	mux.HandleFunc(
		"PATCH /api/courier/orders/{id}/decline",
		h.DeclineAssignment,
	)

	mux.HandleFunc(
		"PATCH /api/courier/orders/{id}/status",
		h.UpdateOrderStatus,
	)
}

// authenticatedCourierID gets the ID of the logged-in user
// from the authentication middleware's request context.
func (h *Handler) authenticatedCourierID(
	w http.ResponseWriter,
	r *http.Request,
) (int64, bool) {

	if h.getUserID == nil {
		writeError(
			w,
			http.StatusInternalServerError,
			"authentication resolver is not configured",
		)
		return 0, false
	}

	id, err := h.getUserID(r)

	if err != nil || id <= 0 {
		writeError(
			w,
			http.StatusUnauthorized,
			"authentication required",
		)
		return 0, false
	}

	return id, true
}

// parseOrderID reads the order ID from the URL.
func parseOrderID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(
		r.PathValue("id"),
		10,
		64,
	)

	if err != nil || id <= 0 {
		return 0, ErrInvalidInput
	}

	return id, nil
}

// GET /api/courier/orders
func (h *Handler) GetAssignedOrders(
	w http.ResponseWriter,
	r *http.Request,
) {

	courierID, ok := h.authenticatedCourierID(w, r)
	if !ok {
		return
	}

	orders, err := h.service.GetAssignedOrders(
		r.Context(),
		courierID,
	)

	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Message: "Assigned orders retrieved successfully",
		Data:    orders,
	})
}

// GET /api/courier/orders/{id}
func (h *Handler) GetOrderDetails(
	w http.ResponseWriter,
	r *http.Request,
) {

	courierID, ok := h.authenticatedCourierID(w, r)
	if !ok {
		return
	}

	orderID, err := parseOrderID(r)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	order, err := h.service.GetOrderDetails(
		r.Context(),
		courierID,
		orderID,
	)

	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Message: "Order details retrieved successfully",
		Data:    order,
	})
}

// PATCH /api/courier/orders/{id}/accept
func (h *Handler) AcceptAssignment(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.respondToAssignment(w, r, true)
}

// PATCH /api/courier/orders/{id}/decline
func (h *Handler) DeclineAssignment(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.respondToAssignment(w, r, false)
}

func (h *Handler) respondToAssignment(
	w http.ResponseWriter,
	r *http.Request,
	accept bool,
) {

	courierID, ok := h.authenticatedCourierID(w, r)
	if !ok {
		return
	}

	orderID, err := parseOrderID(r)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	if accept {
		err = h.service.AcceptAssignment(
			r.Context(),
			courierID,
			orderID,
		)
	} else {
		err = h.service.DeclineAssignment(
			r.Context(),
			courierID,
			orderID,
		)
	}

	if err != nil {
		writeServiceError(w, err)
		return
	}

	message := "Assignment declined successfully"

	if accept {
		message = "Assignment accepted successfully"
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Message: message,
	})
}

// PATCH /api/courier/orders/{id}/status
func (h *Handler) UpdateOrderStatus(
	w http.ResponseWriter,
	r *http.Request,
) {

	courierID, ok := h.authenticatedCourierID(w, r)
	if !ok {
		return
	}

	orderID, err := parseOrderID(r)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	// Limit the request body to 1 MB.
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req CourierStatusUpdateRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid JSON request body",
		)
		return
	}

	// Reject additional JSON values after the first object.
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeError(
			w,
			http.StatusBadRequest,
			"request must contain one JSON object",
		)
		return
	}

	err = h.service.UpdateOrderStatus(
		r.Context(),
		courierID,
		orderID,
		req.Status,
		req.Notes,
	)

	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Message: "Order status updated successfully",
	})
}

// Maps service and repository errors to HTTP responses.
func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid request")

	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "order or assignment not found")

	case errors.Is(err, ErrInvalidAssignment),
		errors.Is(err, ErrInvalidTransition):
		writeError(w, http.StatusConflict, err.Error())

	default:
		// Do not expose internal database errors to clients.
		writeError(
			w,
			http.StatusInternalServerError,
			"an internal server error occurred",
		)
	}
}

func writeError(
	w http.ResponseWriter,
	status int,
	message string,
) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	data any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}
