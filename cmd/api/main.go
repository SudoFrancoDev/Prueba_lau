package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/bia-energy/consumption-service/internal/handler"
	"github.com/bia-energy/consumption-service/internal/middleware"
	"github.com/bia-energy/consumption-service/internal/repository"
	"github.com/bia-energy/consumption-service/internal/service"
	"github.com/bia-energy/consumption-service/pkg/database"
)

func main() {
	dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "3306"))

	// XAMPP defaults: root user, no password, port 3306
	dbCfg := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     dbPort,
		User:     getEnv("DB_USER", "root"),
		Password: getEnv("DB_PASSWORD", ""),
		DBName:   getEnv("DB_NAME", "bia_energy"),
	}

	db, err := database.Connect(dbCfg)
	if err != nil {
		log.Fatalf("cannot connect to database: %v", err)
	}
	defer db.Close()
	log.Println("connected to MySQL")

	// Wire up dependencies (clean architecture: repo → service → handler)
	consumptionRepo := repository.NewConsumptionRepository(db)
	addressSvc := repository.NewMockAddressService()
	consumptionSvc := service.NewConsumptionService(consumptionRepo, addressSvc)
	consumptionHandler := handler.NewConsumptionHandler(consumptionSvc)

	mux := http.NewServeMux()
	consumptionHandler.RegisterRoutes(mux)

	chain := middleware.Logger(middleware.Recovery(mux))

	port := getEnv("PORT", "8080")
	log.Printf("server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, chain); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
