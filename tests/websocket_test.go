package handlers_test

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/websocket"

	"solano-wx/src/cache"
	"solano-wx/src/handlers"
)

func TestWS_ConexaoAceita(t *testing.T) {
	originalCityFn, originalWeatherFn := installWebsocketMocks(t)
	defer func() {
		handlers.SetClimaDependencies(originalCityFn, originalWeatherFn)
	}()

	server := httptest.NewServer(handlers.NewWebSocketHandler(cache.New(time.Minute), time.Minute))
	defer server.Close()

	wsURL := strings.Replace(server.URL, "http://", "ws://", 1) + "/api/v1/ws/clima/Fortaleza"
	ws, err := websocket.Dial(wsURL, "", server.URL)
	assert.NoError(t, err)
	if ws != nil {
		defer ws.Close()
	}

	assert.NotNil(t, ws)
}

func TestWS_PayloadCampos(t *testing.T) {
	originalCityFn, originalWeatherFn := installWebsocketMocks(t)
	defer func() {
		handlers.SetClimaDependencies(originalCityFn, originalWeatherFn)
	}()

	server := httptest.NewServer(handlers.NewWebSocketHandler(cache.New(time.Minute), time.Minute))
	defer server.Close()

	wsURL := strings.Replace(server.URL, "http://", "ws://", 1) + "/api/v1/ws/clima/Fortaleza"
	ws, err := websocket.Dial(wsURL, "", server.URL)
	assert.NoError(t, err)
	defer ws.Close()

	var payload handlers.WSPayload
	err = websocket.JSON.Receive(ws, &payload)
	assert.NoError(t, err)
	assert.Equal(t, "clima_atualizado", payload.Evento)
	assert.NotEmpty(t, payload.Cidade)
	assert.NotEmpty(t, payload.Timestamp)
	assert.NotZero(t, payload.Dados.TempAtual)
}

func TestWS_CidadeNormalizada(t *testing.T) {
	originalCityFn, originalWeatherFn := installWebsocketMocks(t)
	defer func() {
		handlers.SetClimaDependencies(originalCityFn, originalWeatherFn)
	}()

	server := httptest.NewServer(handlers.NewWebSocketHandler(cache.New(time.Minute), time.Minute))
	defer server.Close()

	wsURL := strings.Replace(server.URL, "http://", "ws://", 1) + "/api/v1/ws/clima/fortaleza"
	ws, err := websocket.Dial(wsURL, "", server.URL)
	assert.NoError(t, err)
	defer ws.Close()

	var payload handlers.WSPayload
	err = websocket.JSON.Receive(ws, &payload)
	assert.NoError(t, err)
	assert.Equal(t, "fortaleza", strings.ToLower(payload.Cidade))
}

func installWebsocketMocks(t *testing.T) (func(string) (*handlers.CityData, error), func(float64, float64) (*handlers.WeatherData, error)) {
	t.Helper()

	originalCityFn := handlers.BuscarCidadeFn
	originalWeatherFn := handlers.BuscarClimaFn
	handlers.SetClimaDependencies(
		func(nome string) (*handlers.CityData, error) {
			return &handlers.CityData{
				Nome:      "Fortaleza",
				UF:        "CE",
				Latitude:  -3.71,
				Longitude: -38.54,
			}, nil
		},
		func(lat, lon float64) (*handlers.WeatherData, error) {
			return &handlers.WeatherData{
				TempAtual: 29.5,
				TempMin:   24.0,
				TempMax:   32.0,
				Condicao:  "Ensolarado",
				Umidade:   65,
				VentoKmh:  18.0,
			}, nil
		},
	)
	return originalCityFn, originalWeatherFn
}
