
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

	log.Fatal(http.ListenAndServe(":8080", handler))
}
