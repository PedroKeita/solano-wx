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

func TestClima_Sucesso200(t *testing.T) {
	originalCityFn, originalWeatherFn := installClimaMocks(t,
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
		})
	defer func() {
		handlers.SetClimaDependencies(originalCityFn, originalWeatherFn)
	}()

	c := cache.New(time.Minute)
	handler := handlers.NewClimaHandler(c, time.Minute)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/clima/Fortaleza", nil)
	request.SetPathValue("cidade", "Fortaleza")

	handler(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var payload handlers.ClimaResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &payload)
	assert.NoError(t, err)
	assert.Equal(t, "Fortaleza", payload.Cidade)
	assert.Equal(t, "CE", payload.UF)
	assert.InDelta(t, -3.71, payload.Latitude, 0.0001)
	assert.InDelta(t, -38.54, payload.Longitude, 0.0001)
	assert.InDelta(t, 29.5, payload.Clima.TempAtual, 0.0001)
	assert.InDelta(t, 24.0, payload.Clima.TempMin, 0.0001)
	assert.InDelta(t, 32.0, payload.Clima.TempMax, 0.0001)
	assert.Equal(t, "Ensolarado", payload.Clima.Condicao)
	assert.Equal(t, 65, payload.Clima.Umidade)
	assert.InDelta(t, 18.0, payload.Clima.VentoKmh, 0.0001)
}

func TestClima_CacheHitNaoChamaAPI(t *testing.T) {
	var cityCalls int
	var weatherCalls int
	originalCityFn, originalWeatherFn := installClimaMocks(t,
		func(nome string) (*handlers.CityData, error) {
			cityCalls++
			return &handlers.CityData{
				Nome:      "Fortaleza",
				UF:        "CE",
				Latitude:  -3.71,
				Longitude: -38.54,
			}, nil
		},
		func(lat, lon float64) (*handlers.WeatherData, error) {
			weatherCalls++
			return &handlers.WeatherData{
				TempAtual: 29.5,
				TempMin:   24.0,
				TempMax:   32.0,
				Condicao:  "Ensolarado",
				Umidade:   65,
				VentoKmh:  18.0,
			}, nil
		})
	defer func() {
		handlers.SetClimaDependencies(originalCityFn, originalWeatherFn)
	}()

	c := cache.New(time.Minute)
	handler := handlers.NewClimaHandler(c, time.Minute)

	firstRecorder := httptest.NewRecorder()
	firstRequest := httptest.NewRequest(http.MethodGet, "/api/v1/clima/Fortaleza", nil)
	firstRequest.SetPathValue("cidade", "Fortaleza")
	handler(firstRecorder, firstRequest)

	secondRecorder := httptest.NewRecorder()
	secondRequest := httptest.NewRequest(http.MethodGet, "/api/v1/clima/Fortaleza", nil)
	secondRequest.SetPathValue("cidade", "Fortaleza")
	handler(secondRecorder, secondRequest)

	assert.Equal(t, 1, cityCalls)
	assert.Equal(t, 1, weatherCalls)
	assert.Equal(t, http.StatusOK, firstRecorder.Code)
	assert.Equal(t, http.StatusOK, secondRecorder.Code)
}

func TestClima_NomeInvalido400(t *testing.T) {
	originalCityFn, originalWeatherFn := installClimaMocks(t,
		func(nome string) (*handlers.CityData, error) {
			t.Fatalf("não deveria chamar brasilapi")
			return nil, nil
		},
		func(lat, lon float64) (*handlers.WeatherData, error) {
			t.Fatalf("não deveria chamar clima")
			return nil, nil
		})
	defer func() {
		handlers.SetClimaDependencies(originalCityFn, originalWeatherFn)
	}()

	for _, city := range []string{"123", "X", "", "  "} {
		c := cache.New(time.Minute)
		handler := handlers.NewClimaHandler(c, time.Minute)
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/api/v1/clima/test", nil)
		request.SetPathValue("cidade", city)

		handler(recorder, request)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		var payload map[string]string
		err := json.Unmarshal(recorder.Body.Bytes(), &payload)
		assert.NoError(t, err)
		assert.Equal(t, "nome de cidade inválido", payload["error"])
	}
}

func TestClima_CidadeNaoEncontrada404(t *testing.T) {
	originalCityFn, originalWeatherFn := installClimaMocks(t,
		func(nome string) (*handlers.CityData, error) {
			return nil, handlers.ErrCidadeNaoEncontrada
		},
		func(lat, lon float64) (*handlers.WeatherData, error) {
			t.Fatalf("não deveria chamar clima")
			return nil, nil
		})
	defer func() {
		handlers.SetClimaDependencies(originalCityFn, originalWeatherFn)
	}()

	c := cache.New(time.Minute)
	handler := handlers.NewClimaHandler(c, time.Minute)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/clima/Blumenfeld", nil)
	request.SetPathValue("cidade", "Blumenfeld")

	handler(recorder, request)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	var payload map[string]string
	err := json.Unmarshal(recorder.Body.Bytes(), &payload)
	assert.NoError(t, err)
	assert.Equal(t, "cidade não encontrada", payload["error"])
}

func TestClima_ServicoDown503(t *testing.T) {
	originalCityFn, originalWeatherFn := installClimaMocks(t,
		func(nome string) (*handlers.CityData, error) {
			return nil, assert.AnError
		},
		func(lat, lon float64) (*handlers.WeatherData, error) {
			t.Fatalf("não deveria chamar clima")
			return nil, nil
		})
	defer func() {
		handlers.SetClimaDependencies(originalCityFn, originalWeatherFn)
	}()

	c := cache.New(time.Minute)
	handler := handlers.NewClimaHandler(c, time.Minute)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/clima/Fortaleza", nil)
	request.SetPathValue("cidade", "Fortaleza")

	handler(recorder, request)

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	var payload map[string]string
	err := json.Unmarshal(recorder.Body.Bytes(), &payload)
	assert.NoError(t, err)
	assert.Equal(t, "serviço climático indisponível", payload["error"])
}

func installClimaMocks(t *testing.T, cityFn func(string) (*handlers.CityData, error), climaFn func(float64, float64) (*handlers.WeatherData, error)) (func(string) (*handlers.CityData, error), func(float64, float64) (*handlers.WeatherData, error)) {
	t.Helper()

	originalCityFn := handlers.BuscarCidadeFn
	originalWeatherFn := handlers.BuscarClimaFn
	handlers.BuscarCidadeFn = cityFn
	handlers.BuscarClimaFn = climaFn
	return originalCityFn, originalWeatherFn
}
