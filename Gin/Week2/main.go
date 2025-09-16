package main

import (
	"week2/config"
	"week2/models"
	"week2/routes"
)

func main() {
	config.ConnectDB()

	// Tự động migrate DB
	config.DB.AutoMigrate(&models.User{}, &models.Task{})

	r := routes.SetupRouter()
	r.Run(":8080") // chạy server ở http://localhost:8080
}
