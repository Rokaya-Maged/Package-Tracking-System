package orders

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type OrderHandler struct {
	Service *OrderService
}

// Admin
func (h *OrderHandler) GetAllOrdersHandler(res http.ResponseWriter, req *http.Request) {
	orders, err := h.Service.GetAllOrders()

	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the orders as JSON
	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(orders)
}

// Admin
func (h *OrderHandler) GetOrderByIDHandler(res http.ResponseWriter, req *http.Request) {
	idString := req.PathValue("id")
	id, err := strconv.Atoi(idString)

	if err != nil {
		http.Error(res, "Invalid order ID", http.StatusBadRequest)
		return
	}

	if id <= 0 {
		http.Error(res, "Invalid order ID", http.StatusBadRequest)
		return
	}

	order, err := h.Service.GetOrderByID(id)

	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			http.Error(res, "Order not found", http.StatusNotFound)
			return
		}

		http.Error(res, "Internal server error", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(order)
}

// User
func (h *OrderHandler) CreateOrderHandler(res http.ResponseWriter, req *http.Request) {
	var order Order

	err := json.NewDecoder(req.Body).Decode(&order)
	if err != nil {
		http.Error(res, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.Service.AddOrder(&order)
	if err != nil {
		switch {
		case errors.Is(err, ErrBusinessUserNotFound):
			http.Error(res, err.Error(), http.StatusBadRequest)

		case errors.Is(err, ErrTrackingNumberExists):
			http.Error(res, err.Error(), http.StatusConflict)

		case errors.Is(err, ErrInvalidOrderData):
			http.Error(res, err.Error(), http.StatusBadRequest)

		default:
			http.Error(res, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	json.NewEncoder(res).Encode(order)
}

func (h *OrderHandler) UpdateOrderHandler(res http.ResponseWriter, req *http.Request) {
	idString := req.PathValue("id")
	id, err := strconv.Atoi(idString)

	if err != nil {
		http.Error(res, "Invalid order ID", http.StatusBadRequest)
		return
	}

	if id <= 0 {
		http.Error(res, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var statusUpdate struct {
		Status string `json:"status"`
	}

	err = json.NewDecoder(req.Body).Decode(&statusUpdate)

	if err != nil {
		http.Error(res, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.Service.UpdateOrderStatus(id, statusUpdate.Status)

	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			http.Error(res, "Order not found", http.StatusNotFound)
			return
		}

		if errors.Is(err, ErrInvalidOrderStatus) {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(res, "Internal server error", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	json.NewEncoder(res).Encode(map[string]string{
		"message": "Order status updated successfully",
	})
}

func (h *OrderHandler) DeleteOrderHandler(res http.ResponseWriter, req *http.Request) {
	idString := req.PathValue("id")
	id, err := strconv.Atoi(idString)

	if err != nil {
		http.Error(res, "Invalid order ID", http.StatusBadRequest)
		return
	}

	if id <= 0 {
		http.Error(res, "Invalid order ID", http.StatusBadRequest)
		return
	}

	err = h.Service.DeleteOrder(id)

	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			http.Error(res, "Order not found", http.StatusNotFound)
			return
		}

		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	json.NewEncoder(res).Encode(map[string]string{
		"message": "Order deleted successfully",
	})
}
