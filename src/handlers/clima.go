package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode"

	"solano-wx/src/cache"
	"solano-wx/src/services"
)

var depsMu sync.RWMutex

var (
	ErrCidadeNaoEncontrada = services.ErrCidadeNaoEncontrada
	ErrServicoIndisponivel = services.ErrServicoIndisponivel
)

type CityData struct {
	Nome      string  `json:"nome"`
	UF        string  `json:"uf"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type WeatherData struct {
	TempAtual float64 `json:"temperatura_atual"`
	TempMin   float64 `json:"temperatura_min"`
	TempMax   float64 `json:"temperatura_max"`
	Condicao  string  `json:"condicao"`
	Umidade   int     `json:"umidade"`
	VentoKmh  float64 `json:"vento_kmh"`
}

type ClimaResponse struct {
	Cidade    string      `json:"cidade"`
	UF        string      `json:"uf"`
	Latitude  float64     `json:"latitude"`
	Longitude float64     `json:"longitude"`
	Clima     WeatherData `json:"clima"`
}

type cidadeLookupFunc func(string) (*CityData, error)
type climaLookupFunc func(float64, float64) (*WeatherData, error)

var (
	BuscarCidadeFn cidadeLookupFunc = func(nome string) (*CityData, error) {
		city, err := services.BuscarCidade(nome)
		if err != nil {
			return nil, err
		}
		return &CityData{
			Nome:      city.Nome,
			UF:        city.UF,
			Latitude:  city.Latitude,
			Longitude: city.Longitude,
		}, nil
	}
	BuscarClimaFn climaLookupFunc = func(lat, lon float64) (*WeatherData, error) {
		weather, err := services.BuscarClima(lat, lon)
		if err != nil {
			return nil, err
		}
		return &WeatherData{
			TempAtual: weather.TempAtual,
			TempMin:   weather.TempMin,
			TempMax:   weather.TempMax,
			Condicao:  weather.Condicao,
			Umidade:   weather.Umidade,
			VentoKmh:  weather.VentoKmh,
		}, nil
	}
)

func SetClimaDependencies(cityFn cidadeLookupFunc, climaFn climaLookupFunc) {
	depsMu.Lock()
	if cityFn != nil {
		BuscarCidadeFn = cityFn
	}
	if climaFn != nil {
		BuscarClimaFn = climaFn
	}
	depsMu.Unlock()
}

// @Summary     Clima atual de uma cidade
// @Description Retorna dados climáticos e geográficos. Usa cache de 10min.
// @Tags        clima
// @Produce     json
// @Param       cidade  path  string  true  "Nome da cidade brasileira"
// @Success     200  {object}  ClimaResponse
// @Failure     400  {object}  models.ErrorResponse
// @Failure     404  {object}  models.ErrorResponse
// @Failure     503  {object}  models.ErrorResponse
// @Router      /clima/{cidade} [get]
func NewClimaHandler(c *cache.Cache, ttlClima time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		depsMu.RLock()
		cityFn := BuscarCidadeFn
		climaFn := BuscarClimaFn
		depsMu.RUnlock()

		cidade := strings.TrimSpace(r.PathValue("cidade"))
		if isInvalidCidade(cidade) {
			writeJSONError(w, http.StatusBadRequest, "nome de cidade inválido")
			return
		}

		normalized := normalizeCidade(cidade)
		cacheKey := "clima:" + normalized

		if c != nil {
			if cached, ok := c.Get(cacheKey); ok {
				if response, ok := cached.(ClimaResponse); ok {
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(response)
					return
				}
			}
		}

		city, err := cityFn(cidade)
		if err != nil {
			if errors.Is(err, ErrCidadeNaoEncontrada) {
				writeJSONError(w, http.StatusNotFound, "cidade não encontrada")
				return
			}

			writeJSONError(w, http.StatusServiceUnavailable, "serviço climático indisponível")
			return
		}

		weather, err := climaFn(city.Latitude, city.Longitude)
		if err != nil {
			writeJSONError(w, http.StatusServiceUnavailable, "serviço climático indisponível")
			return
		}

		response := ClimaResponse{
			Cidade:    city.Nome,
			UF:        city.UF,
			Latitude:  city.Latitude,
			Longitude: city.Longitude,
			Clima: WeatherData{
				TempAtual: weather.TempAtual,
				TempMin:   weather.TempMin,
				TempMax:   weather.TempMax,
				Condicao:  weather.Condicao,
				Umidade:   weather.Umidade,
				VentoKmh:  weather.VentoKmh,
			},
		}

		if c != nil {
			c.Set(cacheKey, response, ttlClima)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}

func normalizeCidade(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	replacer := strings.NewReplacer(
		"á", "a",
		"à", "a",
		"â", "a",
		"ã", "a",
		"ä", "a",
		"Á", "a",
		"À", "a",
		"Â", "a",
		"Ã", "a",
		"Ä", "a",
		"é", "e",
		"è", "e",
		"ê", "e",
		"ë", "e",
		"É", "e",
		"È", "e",
		"Ê", "e",
		"Ë", "e",
		"í", "i",
		"ì", "i",
		"î", "i",
		"ï", "i",
		"Í", "i",
		"Ì", "i",
		"Î", "i",
		"Ï", "i",
		"ó", "o",
		"ò", "o",
		"ô", "o",
		"õ", "o",
		"ö", "o",
		"Ó", "o",
		"Ò", "o",
		"Ô", "o",
		"Õ", "o",
		"Ö", "o",
		"ú", "u",
		"ù", "u",
		"û", "u",
		"ü", "u",
		"Ú", "u",
		"Ù", "u",
		"Û", "u",
		"Ü", "u",
		"ç", "c",
		"Ç", "c",
	)
	return replacer.Replace(s)
}

func isInvalidCidade(s string) bool {
	if len(s) < 2 {
		return true
	}

	hasLetter := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			hasLetter = true
			break
		}
	}

	return !hasLetter
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
