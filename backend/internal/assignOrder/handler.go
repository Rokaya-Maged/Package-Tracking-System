package assignOrder

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type AssignmentHandler struct {
	Service *AssignmentService
}

func handleAssignmentError(res http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidOrderID),
		errors.Is(err, ErrInvalidCourierID),
		errors.Is(err, ErrInvalidAssignedBy):
		http.Error(res, err.Error(), http.StatusBadRequest)

	case errors.Is(err, ErrUserNotCourier),
		errors.Is(err, ErrUserNotAdmin):
		http.Error(res, err.Error(), http.StatusForbidden)

	case errors.Is(err, ErrUserNotFound),
		errors.Is(err, ErrOrderNotFound),
		errors.Is(err, ErrNoActiveAssignment):
		http.Error(res, err.Error(), http.StatusNotFound)

	case errors.Is(err, ErrOrderAlreadyAssigned),
		errors.Is(err, ErrOrderNotPending),
		errors.Is(err, ErrSameCourier):
		http.Error(res, err.Error(), http.StatusConflict)

	default:
		http.Error(res, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *AssignmentHandler) AssignOrderToCourierHandler(res http.ResponseWriter, req *http.Request) {
	var request struct {
		OrderID    int `json:"order_id"`
		CourierID  int `json:"courier_id"`
		AssignedBy int `json:"assigned_by"`
	}

	err := json.NewDecoder(req.Body).Decode(&request)

	if err != nil {
		http.Error(res, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = h.Service.AssignOrderToCourier(request.OrderID, request.CourierID, request.AssignedBy)
	if err != nil {
		handleAssignmentError(res, err)
		return
	}

	res.WriteHeader(http.StatusCreated)
}

func (h *AssignmentHandler) ReassignOrderToCourierHandler(res http.ResponseWriter, req *http.Request) {
	var request struct {
		OrderID      int `json:"order_id"`
		NewCourierID int `json:"new_courier_id"`
		AssignedBy   int `json:"assigned_by"`
	}

	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		http.Error(res, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = h.Service.ReassignOrderToCourier(request.OrderID, request.NewCourierID, request.AssignedBy)
	if err != nil {
		handleAssignmentError(res, err)
		return
	}

	res.WriteHeader(http.StatusOK)
}

// get all assignments handler
func (h *AssignmentHandler) GetAllAssignmetHandler(res http.ResponseWriter, req *http.Request) {
	assignments, err := h.Service.GetAllAssignments()
	if err != nil {
		handleAssignmentError(res, err)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(assignments)
}

// get assignments by order id handler
func (h *AssignmentHandler) GetAssignmentsByOrderID(res http.ResponseWriter, req *http.Request) {

	idString := req.PathValue("order_id")
	id, err := strconv.Atoi(idString)

	if err != nil || id <= 0 {
		http.Error(res, "Invalid order ID", http.StatusBadRequest)
		return
	}

	assignments, err := h.Service.GetAssignmentsByOrderID(id)

	if err != nil {
		handleAssignmentError(res, err)
		return
	}

	// Return assignments as JSON
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	json.NewEncoder(res).Encode(assignments)
}

// get assignments by courier id handler
func (h *AssignmentHandler) GetAssignmentsByCourierID(res http.ResponseWriter, req *http.Request) {
	idString := req.PathValue("courier_id")
	id, err := strconv.Atoi(idString)

	if err != nil || id <= 0 {
		http.Error(res, "Invalid courier ID", http.StatusBadRequest)
		return
	}

	assignments, err := h.Service.GetAssignmentsByCourierID(id)

	if err != nil {
		handleAssignmentError(res, err)
		return
	}

	// Return assignments as JSON
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	json.NewEncoder(res).Encode(assignments)
}

func (h *AssignmentHandler) GetAssignmentsByStatus(res http.ResponseWriter, req *http.Request) {

	status := req.URL.Query().Get("status")

	if strings.TrimSpace(status) == "" {
		http.Error(res, "Status cannot be empty", http.StatusBadRequest)
		return
	}

	assignments, err := h.Service.GetAssignmentsByStatus(status)
	if err != nil {
		handleAssignmentError(res, err)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(assignments)
}
