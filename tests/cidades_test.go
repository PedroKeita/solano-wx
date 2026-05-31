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

func TestCidades_Sucesso200(t *testing.T) {
	originalFn := installCidadesMock(t, func(uf string) ([]string, error) {
		assert.Equal(t, "CE", uf)
		return []string{"Fortaleza", "Caucaia", "Sobral"}, nil
	})
	defer func() { handlers.SetCidadesDependencies(originalFn) }()

	handler := handlers.NewCidadesHandler(cache.New(time.Hour), 24*time.Hour)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/cidades/CE", nil)
	request.SetPathValue("uf", "CE")

	handler(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	var payload handlers.CidadesResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &payload)
	assert.NoError(t, err)
	assert.Equal(t, "CE", payload.UF)
	assert.Equal(t, 3, payload.Total)
	assert.Len(t, payload.Cidades, 3)
	assert.Equal(t, []string{"Fortaleza", "Caucaia", "Sobral"}, payload.Cidades)
}

func TestCidades_ComLimite(t *testing.T) {
	originalFn := installCidadesMock(t, func(uf string) ([]string, error) {
		return []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}, nil
	})
	defer func() { handlers.SetCidadesDependencies(originalFn) }()

	handler := handlers.NewCidadesHandler(cache.New(time.Hour), 24*time.Hour)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/cidades/CE?limite=3", nil)
	request.SetPathValue("uf", "CE")

	handler(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	var payload handlers.CidadesResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &payload)
	assert.NoError(t, err)
	assert.Equal(t, 3, payload.Total)
	assert.Len(t, payload.Cidades, 3)
	assert.Equal(t, []string{"A", "B", "C"}, payload.Cidades)
}

func TestCidades_UFInvalida400(t *testing.T) {
	originalFn := installCidadesMock(t, func(uf string) ([]string, error) {
		t.Fatalf("não deveria chamar mock")
		return nil, nil
	})
	defer func() { handlers.SetCidadesDependencies(originalFn) }()

	for _, uf := range []string{"123", "S", "SPP", "", "R1"} {
		handler := handlers.NewCidadesHandler(cache.New(time.Hour), 24*time.Hour)
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/api/v1/cidades/teste", nil)
		request.SetPathValue("uf", uf)

		handler(recorder, request)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		var payload map[string]string
		err := json.Unmarshal(recorder.Body.Bytes(), &payload)
		assert.NoError(t, err)
		assert.Equal(t, "UF inválida", payload["error"])
	}
}

func TestCidades_UFInexistente404(t *testing.T) {
	originalFn := installCidadesMock(t, func(uf string) ([]string, error) {
		assert.Equal(t, "XX", uf)
		return []string{}, nil
	})
	defer func() { handlers.SetCidadesDependencies(originalFn) }()

	handler := handlers.NewCidadesHandler(cache.New(time.Hour), 24*time.Hour)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/cidades/XX", nil)
	request.SetPathValue("uf", "XX")

	handler(recorder, request)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	var payload map[string]string
	err := json.Unmarshal(recorder.Body.Bytes(), &payload)
	assert.NoError(t, err)
	assert.Equal(t, "estado não encontrado", payload["error"])
}

func TestCidades_CacheHit(t *testing.T) {
	var calls int
	originalFn := installCidadesMock(t, func(uf string) ([]string, error) {
		calls++
		return []string{"Fortaleza", "Caucaia"}, nil
	})
	defer func() { handlers.SetCidadesDependencies(originalFn) }()

	c := cache.New(time.Hour)
	handler := handlers.NewCidadesHandler(c, 24*time.Hour)

	firstRecorder := httptest.NewRecorder()
	firstRequest := httptest.NewRequest(http.MethodGet, "/api/v1/cidades/CE", nil)
	firstRequest.SetPathValue("uf", "CE")
	handler(firstRecorder, firstRequest)

	secondRecorder := httptest.NewRecorder()
	secondRequest := httptest.NewRequest(http.MethodGet, "/api/v1/cidades/CE", nil)
	secondRequest.SetPathValue("uf", "CE")
	handler(secondRecorder, secondRequest)

	assert.Equal(t, 1, calls)
	assert.Equal(t, http.StatusOK, firstRecorder.Code)
	assert.Equal(t, http.StatusOK, secondRecorder.Code)
}

func installCidadesMock(t *testing.T, fn func(string) ([]string, error)) func(string) ([]string, error) {
	t.Helper()
	original := handlers.BuscarCidadesFn
	handlers.BuscarCidadesFn = fn
	return original
}
