package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
	"unicode"

	"solano-wx/src/models"
)

var (
	ErrCidadeNaoEncontrada = errors.New("cidade não encontrada")
	ErrServicoIndisponivel = errors.New("serviço climático indisponível")
)

var localCities = map[string]models.CityData{
	"fortaleza":      {Nome: "Fortaleza", UF: "CE", Latitude: -3.7319, Longitude: -38.5267},
	"caucaia":        {Nome: "Caucaia", UF: "CE", Latitude: -3.7361, Longitude: -38.6531},
	"sobral":         {Nome: "Sobral", UF: "CE", Latitude: -3.6894, Longitude: -40.3480},
	"sao paulo":      {Nome: "São Paulo", UF: "SP", Latitude: -23.5505, Longitude: -46.6333},
	"rio de janeiro": {Nome: "Rio de Janeiro", UF: "RJ", Latitude: -22.9068, Longitude: -43.1729},
	"brasilia":       {Nome: "Brasília", UF: "DF", Latitude: -15.7939, Longitude: -47.8828},
	"salvador":       {Nome: "Salvador", UF: "BA", Latitude: -12.9777, Longitude: -38.5016},
	"recife":         {Nome: "Recife", UF: "PE", Latitude: -8.0476, Longitude: -34.8770},
	"manaus":         {Nome: "Manaus", UF: "AM", Latitude: -3.1190, Longitude: -60.0217},
	"curitiba":       {Nome: "Curitiba", UF: "PR", Latitude: -25.4284, Longitude: -49.2733},
	"porto alegre":   {Nome: "Porto Alegre", UF: "RS", Latitude: -30.0346, Longitude: -51.2177},
	"belo horizonte": {Nome: "Belo Horizonte", UF: "MG", Latitude: -19.9167, Longitude: -43.9345},
}

var localCitiesByUF = map[string][]string{
	"CE": {"Fortaleza", "Caucaia", "Sobral", "Juazeiro do Norte", "Maracanaú"},
	"SP": {"São Paulo", "Campinas", "Santos", "Sorocaba", "Ribeirão Preto"},
	"RJ": {"Rio de Janeiro", "Niterói", "Petrópolis", "Nova Iguaçu", "Duque de Caxias"},
	"DF": {"Brasília", "Ceilândia", "Taguatinga", "Gama", "Sobradinho"},
	"BA": {"Salvador", "Feira de Santana", "Vitória da Conquista", "Ilhéus", "Juazeiro"},
	"PE": {"Recife", "Olinda", "Jaboatão dos Guararapes", "Caruaru", "Petrolina"},
	"MG": {"Belo Horizonte", "Uberlândia", "Contagem", "Juiz de Fora", "Betim"},
	"RS": {"Porto Alegre", "Caxias do Sul", "Pelotas", "Canoas", "Santa Maria"},
	"PR": {"Curitiba", "Londrina", "Maringá", "Ponta Grossa", "Foz do Iguaçu"},
	"AM": {"Manaus", "Parintins", "Itacoatiara", "Manacapuru", "Tefé"},
}

func BuscarCidade(nome string) (*models.CityData, error) {
	normalized := normalize(nome)
	if normalized == "" {
		return nil, ErrCidadeNaoEncontrada
	}

	if city, ok := localCities[normalized]; ok {
		copy := city
		return &copy, nil
	}

	endpoint := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=5&language=pt&format=json&country=BR", url.QueryEscape(strings.TrimSpace(nome)))
	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, ErrServicoIndisponivel
	}

	client := &http.Client{Timeout: 6 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, ErrServicoIndisponivel
	}
	defer response.Body.Close()

	if response.StatusCode >= 500 {
		return nil, ErrServicoIndisponivel
	}

	var payload struct {
		Results []struct {
			Name        string  `json:"name"`
			Admin1      string  `json:"admin1"`
			Latitude    float64 `json:"latitude"`
			Longitude   float64 `json:"longitude"`
			CountryCode string  `json:"country_code"`
		} `json:"results"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, ErrServicoIndisponivel
	}

	for _, result := range payload.Results {
		if strings.EqualFold(strings.TrimSpace(result.CountryCode), "BR") {
			city := models.CityData{Nome: result.Name, Latitude: result.Latitude, Longitude: result.Longitude, UF: result.Admin1}
			return &city, nil
		}
	}

	if len(payload.Results) == 0 {
		return nil, ErrCidadeNaoEncontrada
	}

	result := payload.Results[0]
	city := models.CityData{Nome: result.Name, Latitude: result.Latitude, Longitude: result.Longitude, UF: result.Admin1}
	return &city, nil
}

func ListarCidades(uf string) ([]string, error) {
	uf = strings.ToUpper(strings.TrimSpace(uf))
	if uf == "" {
		return nil, ErrCidadeNaoEncontrada
	}

	if cities, ok := localCitiesByUF[uf]; ok {
		result := make([]string, len(cities))
		copy(result, cities)
		return result, nil
	}

	endpoint := fmt.Sprintf("https://servicodados.ibge.gov.br/api/v1/localidades/estados/%s/municipios", url.PathEscape(uf))
	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, ErrServicoIndisponivel
	}

	client := &http.Client{Timeout: 6 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, ErrServicoIndisponivel
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return []string{}, nil
	}
	if response.StatusCode >= 500 {
		return nil, ErrServicoIndisponivel
	}

	var payload []struct {
		Nome string `json:"nome"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, ErrServicoIndisponivel
	}

	cities := make([]string, 0, len(payload))
	for _, item := range payload {
		if item.Nome != "" {
			cities = append(cities, item.Nome)
		}
	}

	if len(cities) == 0 {
		return []string{}, nil
	}

	sort.Strings(cities)
	return cities, nil
}

func normalize(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(
		"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"í", "i", "ì", "i", "î", "i", "ï", "i",
		"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
		"ú", "u", "ù", "u", "û", "u", "ü", "u",
		"ç", "c",
	)
	return replacer.Replace(value)
}

func CitySampleByCoords(lat, lon float64) (*models.CityData, bool) {
	for _, city := range localCities {
		if approx(city.Latitude, lat) && approx(city.Longitude, lon) {
			copy := city
			return &copy, true
		}
	}
	return nil, false
}

func approx(a, b float64) bool {
	const tolerance = 0.25
	if a > b {
		return a-b <= tolerance
	}
	return b-a <= tolerance
}

func NormalizeLocal(value string) string {
	return normalize(value)
}

func IsBrazilianLetterString(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if !unicode.IsLetter(r) && r != ' ' && r != '-' {
			return false
		}
	}
	return true
}
