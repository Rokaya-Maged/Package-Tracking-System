
package main

import (
    "fmt"
    "log"

    "tracking-system/config"
)

func main() {
    // Connect to PostgreSQL
    err := config.ConnectDB()
    if err != nil {
        log.Fatal("Database connection failed: ", err)
    }

    // Close database connection when application exits
    defer config.DB.Close()

    fmt.Println("Backend is running!")
}
