package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"solano-wx/src/cache"
)

type CidadesResponse struct {
	UF      string   `json:"uf"`
	Total   int      `json:"total"`
	Cidades []string `json:"cidades"`
}

type cidadesLookupFunc func(string) ([]string, error)

var BuscarCidadesFn cidadesLookupFunc = func(string) ([]string, error) {
	return nil, ErrServicoIndisponivel
}

func SetCidadesDependencies(cidadesFn cidadesLookupFunc) {
	if cidadesFn != nil {
		BuscarCidadesFn = cidadesFn
	}
}

var ufRegexp = regexp.MustCompile(`^[A-Z]{2}$`)

// @Summary     Lista municípios de um estado
// @Description Retorna todos os municípios de uma UF. Usa cache de 24h.
// @Tags        cidades
// @Produce     json
// @Param       uf      path   string  true   "Sigla do estado ex: CE"
// @Param       limite  query  int     false  "Número máximo de resultados"
// @Success     200  {object}  CidadesResponse
// @Failure     400  {object}  models.ErrorResponse
// @Failure     404  {object}  models.ErrorResponse
// @Router      /cidades/{uf} [get]
func NewCidadesHandler(c *cache.Cache, ttlGeo time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		uf := strings.ToUpper(strings.TrimSpace(r.PathValue("uf")))
		if !ufRegexp.MatchString(uf) {
			writeJSONError(w, http.StatusBadRequest, "UF inválida")
			return
		}

		cacheKey := "cidades:" + uf
		if c != nil {
			if cached, ok := c.Get(cacheKey); ok {
				if response, ok := cached.(CidadesResponse); ok {
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(response)
					return
				}
			}
		}

		cidades, err := BuscarCidadesFn(uf)
		if err != nil {
			writeJSONError(w, http.StatusServiceUnavailable, "serviço indisponível")
			return
		}
		if len(cidades) == 0 {
			writeJSONError(w, http.StatusNotFound, "estado não encontrado")
			return
		}

		if limiteStr := strings.TrimSpace(r.URL.Query().Get("limite")); limiteStr != "" {
			if limite, err := strconv.Atoi(limiteStr); err == nil && limite > 0 && limite < len(cidades) {
				cidades = cidades[:limite]
			}
		}

		response := CidadesResponse{
			UF:      uf,
			Total:   len(cidades),
			Cidades: cidades,
		}

		if c != nil {
			c.Set(cacheKey, response, ttlGeo)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}
