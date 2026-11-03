package main

import (
	"log"
	"net/http"
	"tracking-system/config"
	"tracking-system/internal/assignOrder"
	"tracking-system/internal/courier"
	"tracking-system/internal/orders"
	"tracking-system/internal/router"
)

// to connect with frontend on different port
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {

	// Connect to Database
	err := config.ConnectDB()
	if err != nil {
		log.Fatal("Database connection failed: ", err)
	}
	defer config.DB.Close()

	// Create Orders Repository
	orderRepo := &orders.OrderRepository{
		DB: config.DB,
	}

	// Create Assignment Repository
	assignmentRepo := &assignOrder.AssignmentRepository{
		DB: config.DB,
	}

	// Create Courier Repository
	courierRepo := courier.NewRepository(config.DB)

	// Create Orders Service
	orderService := &orders.OrderService{
		Repo: orderRepo,
	}

	// Create Assignment Service
	assignmentService := &assignOrder.AssignmentService{
		Repo: assignmentRepo,
	}

	// Create Courier Service
	courierService := courier.NewService(courierRepo)

	// Create Main Router
	handler := router.New(
		orderService,
		assignmentService,
		courierService,
		nil,
	)

	// Start Server
	log.Println("Backend is running on http://localhost:8080")

	log.Fatal(http.ListenAndServe(":8080", enableCORS(handler)))
}
