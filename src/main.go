// @title           solano-wx API
// @version         1.0
// @description     API REST de dados climáticos e geográficos de cidades brasileiras
// @host            localhost:3000
// @BasePath        /api/v1
package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"solano-wx/src/cache"
	"solano-wx/src/handlers"
)

func main() {
	port := getEnvInt("PORT", 3000)
	ttlClima := time.Duration(getEnvInt("CACHE_TTL_CLIMA", 600)) * time.Second
	ttlGeo := time.Duration(getEnvInt("CACHE_TTL_GEO", 86400)) * time.Second

	sharedCache := cache.New(ttlClima)
	startTime := time.Now()

	mux := http.NewServeMux()
	mux.HandleFunc("/dashboard", handlers.NewDashboardHandler())
	mux.Handle("/docs/", handlers.NewDocsHandler())
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/api/v1/health", handlers.NewHealthHandler(sharedCache, startTime))
	mux.Handle("/api/v1/clima/{cidade}", handlers.NewClimaHandler(sharedCache, ttlClima))
	mux.HandleFunc("/api/v1/clima/{cidade}/previsao", handlers.NewPrevisaoHandler())
	mux.Handle("/api/v1/cidades/{uf}", handlers.NewCidadesHandler(sharedCache, ttlGeo))
	mux.Handle("/api/v1/ws/clima/{cidade}", handlers.NewWebSocketHandler(sharedCache, ttlClima))

	server := &http.Server{
		Addr:              ":" + strconv.Itoa(port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("solano-wx listening on :%d", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func getEnvInt(name string, fallback int) int {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}
