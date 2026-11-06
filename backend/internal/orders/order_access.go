package orders

import (
	"net/http"
	"strconv"

	"tracking-system/internal/middleware"
)

func OrderDetailsByRole(service *OrderService, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := middleware.GetUser(r)

		if !ok {
			http.Error(w, "Login required", http.StatusUnauthorized)
			return
		}

		if user.Role == "admin" {
			next.ServeHTTP(w, r)
			return
		}

		if user.Role != "business" {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil || id <= 0 {
			http.Error(w, "Invalid order ID", http.StatusBadRequest)
			return
		}

		order, err := service.GetOrderByID(id)
		if err != nil {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}

		if order.BusinessUserID != user.ID {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
