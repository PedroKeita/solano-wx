package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"solano-wx/src/cache"
	"solano-wx/src/handlers"
)

func TestHealth_Status200(t *testing.T) {
	handler := handlers.NewHealthHandler(cache.New(time.Minute), time.Now())
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)

	handler(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestHealth_CamposObrigatorios(t *testing.T) {
	handler := handlers.NewHealthHandler(cache.New(time.Minute), time.Now())
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)

	handler(recorder, request)

	var payload map[string]any
	err := json.Unmarshal(recorder.Body.Bytes(), &payload)
	assert.NoError(t, err)
	assert.Contains(t, payload, "status")
	assert.Contains(t, payload, "cache")
	assert.Contains(t, payload, "uptime")
	assert.Equal(t, "healthy", payload["status"])

	cachePayload, ok := payload["cache"].(map[string]any)
	assert.True(t, ok)
	assert.Contains(t, cachePayload, "entradas")
	assert.Contains(t, cachePayload, "hits")
	assert.Contains(t, cachePayload, "misses")
	assert.Contains(t, cachePayload, "hit_rate")
}

func TestHealth_UptimeCresce(t *testing.T) {
	handler := handlers.NewHealthHandler(cache.New(time.Minute), time.Now().Add(-5*time.Second))

	firstRecorder := httptest.NewRecorder()
	handler(firstRecorder, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))

	time.Sleep(1100 * time.Millisecond)

	secondRecorder := httptest.NewRecorder()
	handler(secondRecorder, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))

	var firstPayload map[string]any
	var secondPayload map[string]any

	err := json.Unmarshal(firstRecorder.Body.Bytes(), &firstPayload)
	assert.NoError(t, err)
	err = json.Unmarshal(secondRecorder.Body.Bytes(), &secondPayload)
	assert.NoError(t, err)

	firstUptime, ok := firstPayload["uptime"].(string)
	assert.True(t, ok)
	secondUptime, ok := secondPayload["uptime"].(string)
	assert.True(t, ok)

	assert.Contains(t, firstUptime, "s")
	assert.Contains(t, secondUptime, "s")
	assert.NotEqual(t, "0s", firstUptime)
	assert.NotEqual(t, "0s", secondUptime)

	firstDuration, err := time.ParseDuration(firstUptime)
	assert.NoError(t, err)
	secondDuration, err := time.ParseDuration(secondUptime)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, secondDuration, firstDuration)
}
