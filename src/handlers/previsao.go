package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"solano-wx/src/services"
)

type PrevisaoResponse struct {
	Data string  `json:"data"`
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
}

func NewPrevisaoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		cidade := strings.TrimSpace(r.PathValue("cidade"))
		if isInvalidCidade(cidade) {
			writeJSONError(w, http.StatusBadRequest, "nome de cidade inválido")
			return
		}

		city, err := services.BuscarCidade(cidade)
		if err != nil {
			if err == services.ErrCidadeNaoEncontrada {
				writeJSONError(w, http.StatusNotFound, "cidade não encontrada")
				return
			}
			writeJSONError(w, http.StatusServiceUnavailable, "serviço climático indisponível")
			return
		}

		forecast, err := services.BuscarPrevisao(city.Latitude, city.Longitude)
		if err != nil {
			writeJSONError(w, http.StatusServiceUnavailable, "serviço climático indisponível")
			return
		}

		response := make([]PrevisaoResponse, 0, len(forecast))
		for _, day := range forecast {
			response = append(response, PrevisaoResponse{Data: day.Data, Min: day.Min, Max: day.Max})
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}
