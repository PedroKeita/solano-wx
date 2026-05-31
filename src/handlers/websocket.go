package handlers

import (
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/websocket"

	"solano-wx/src/cache"
)

type WSPayload struct {
	Evento    string      `json:"evento"`
	Cidade    string      `json:"cidade"`
	Dados     WeatherData `json:"dados"`
	Timestamp string      `json:"timestamp"`
}

func NewWebSocketHandler(c *cache.Cache, ttlClima time.Duration) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/v1/ws/clima/{cidade}", websocket.Handler(func(ws *websocket.Conn) {
		defer func() {
			_ = ws.Close()
		}()

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		cidade := strings.TrimSpace(ws.Request().PathValue("cidade"))
		if cidade == "" {
			return
		}

		if err := sendWeatherSnapshot(ws, cidade, c, ttlClima); err != nil {
			return
		}

		for range ticker.C {
			if err := sendWeatherSnapshot(ws, cidade, c, ttlClima); err != nil {
				return
			}
		}
	}))

	return mux
}

func buscarClimaAtual(cidade string, c *cache.Cache, ttl time.Duration) (*WeatherData, string, error) {
	cacheKey := "clima:" + normalizeCidade(cidade)

	if c != nil {
		if cached, ok := c.Get(cacheKey); ok {
			if response, ok := cached.(ClimaResponse); ok {
				weather := response.Clima
				return &weather, response.Cidade, nil
			}
		}
	}

	city, err := BuscarCidadeFn(cidade)
	if err != nil {
		if err == ErrCidadeNaoEncontrada {
			return nil, "", err
		}
		return nil, "", ErrServicoIndisponivel
	}

	weather, err := BuscarClimaFn(city.Latitude, city.Longitude)
	if err != nil {
		return nil, "", ErrServicoIndisponivel
	}

	response := ClimaResponse{
		Cidade:    city.Nome,
		UF:        city.UF,
		Latitude:  city.Latitude,
		Longitude: city.Longitude,
		Clima:     *weather,
	}

	if c != nil {
		c.Set(cacheKey, response, ttl)
	}

	return weather, city.Nome, nil
}

func sendWeatherSnapshot(ws *websocket.Conn, cidade string, c *cache.Cache, ttl time.Duration) error {
	weather, nomeReal, err := buscarClimaAtual(cidade, c, ttl)
	if err != nil {
		return err
	}

	payload := WSPayload{
		Evento:    "clima_atualizado",
		Cidade:    nomeReal,
		Dados:     *weather,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if err := websocket.JSON.Send(ws, payload); err != nil {
		return err
	}

	return nil
}
