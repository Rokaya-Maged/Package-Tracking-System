package orders

import (
	"encoding/json"
	"net/http"

	"tracking-system/internal/middleware"
)

func OrdersListByRole(service *OrderService, adminHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := middleware.GetUser(r)

		if !ok {
			http.Error(w, "Login required", http.StatusUnauthorized)
			return
		}

		if user.Role == "admin" {
			adminHandler.ServeHTTP(w, r)
			return
		}

		if user.Role != "business" {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		allOrders, err := service.GetAllOrders()
		if err != nil {
			http.Error(w, "Failed to load orders", http.StatusInternalServerError)
			return
		}

		myOrders := []Order{}

		for _, order := range allOrders {
			if order.BusinessUserID == user.ID {
				myOrders = append(myOrders, order)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(myOrders)
	})
}
