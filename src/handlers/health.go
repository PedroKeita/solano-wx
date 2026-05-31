package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"solano-wx/src/cache"
)

type HealthResponse struct {
	Status string           `json:"status"`
	Cache  cache.CacheStats `json:"cache"`
	Uptime string           `json:"uptime"`
}

// @Summary     Health check da API
// @Description Retorna status, estatísticas do cache e uptime
// @Tags        infra
// @Produce     json
// @Success     200  {object}  HealthResponse
// @Router      /health [get]
func NewHealthHandler(c *cache.Cache, startTime time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := HealthResponse{
			Status: "healthy",
			Cache:  c.Stats(),
			Uptime: time.Since(startTime).Round(time.Second).String(),
		}

		_ = json.NewEncoder(w).Encode(response)
	}
}
