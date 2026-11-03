
package courier

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type UserIDResolver func(*http.Request) (int64, error)

type Handler struct {
	Service       *Service
	ResolveUserID UserIDResolver
}

func NewHandler(service *Service, resolver UserIDResolver) *Handler {
	return &Handler{
		Service:       service,
		ResolveUserID: resolver,
	}
}

// RegisterRoutes registers all Courier endpoints.
// The supplied mux should be wrapped with authentication and courier-role
// middleware when it is mounted in the main application router.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/courier/orders", h.GetAssignedOrders)
	mux.HandleFunc("GET /api/courier/orders/{id}", h.GetOrderDetails)
	mux.HandleFunc("PATCH /api/courier/orders/{id}/accept", h.AcceptAssignment)
	mux.HandleFunc("PATCH /api/courier/orders/{id}/decline", h.DeclineAssignment)
	mux.HandleFunc("PATCH /api/courier/orders/{id}/status", h.UpdateOrderStatus)
}

func (h *Handler) authenticatedCourierID(
	w http.ResponseWriter,
	r *http.Request,
) (int64, bool) {

	if h.ResolveUserID == nil {
		writeError(w, http.StatusInternalServerError, "authentication resolver is not configured")
		return 0, false
	}

	userID, err := h.ResolveUserID(r)
	if err != nil || userID <= 0 {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return 0, false
	}

	return userID, true
}

func (h *Handler) GetAssignedOrders(
	w http.ResponseWriter,
	r *http.Request,
) {
	courierID, ok := h.authenticatedCourierID(w, r)
	if !ok {
		return
	}

	orders, err := h.Service.GetAssignedOrders(r.Context(), courierID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Message: "Assigned orders retrieved successfully",
		Data:    orders,
	})
}

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
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	order, err := h.Service.GetOrderDetails(
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

func (h *Handler) AcceptAssignment(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.respondToAssignment(w, r, true)
}

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
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if accept {
		err = h.Service.AcceptAssignment(r.Context(), courierID, orderID)
	} else {
		err = h.Service.DeclineAssignment(r.Context(), courierID, orderID)
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
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var request CourierStatusUpdateRequest

	decoder := json.NewDecoder(
		http.MaxBytesReader(w, r.Body, 1<<20),
	)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		writeError(w, http.StatusBadRequest, "request must contain only one JSON object")
		return
	}

	err = h.Service.UpdateOrderStatus(
		r.Context(),
		courierID,
		orderID,
		request.Status,
		request.Notes,
	)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Message: "Order status updated successfully",
	})
}

func parseOrderID(r *http.Request) (int64, error) {
	rawID := strings.TrimSpace(r.PathValue("id"))

	orderID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || orderID <= 0 {
		return 0, errors.New("invalid order ID")
	}

	return orderID, nil
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())

	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "assignment or order not found")

	case errors.Is(err, ErrInvalidAssignment),
		errors.Is(err, ErrInvalidTransition),
		errors.Is(err, ErrNotAccepted),
		errors.Is(err, ErrOrderNotAssignable):
		writeError(w, http.StatusConflict, err.Error())

	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, APIResponse{
		Message: message,
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
