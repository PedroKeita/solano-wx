package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"solano-wx/src/models"
)

func BuscarClima(lat, lon float64) (*models.WeatherData, error) {
	if city, ok := CitySampleByCoords(lat, lon); ok {
		return sampleWeatherForCity(city.Nome), nil
	}

	data, err := fetchForecast(lat, lon)
	if err != nil {
		return nil, ErrServicoIndisponivel
	}

	return &models.WeatherData{
		TempAtual: data.currentTemp,
		TempMin:   data.tempMin,
		TempMax:   data.tempMax,
		Condicao:  weatherCodeToText(data.weatherCode),
		Umidade:   data.humidity,
		VentoKmh:  data.windSpeed,
	}, nil
}

func BuscarPrevisao(lat, lon float64) ([]models.ForecastDay, error) {
	if city, ok := CitySampleByCoords(lat, lon); ok {
		return sampleForecastForCity(city.Nome), nil
	}

	data, err := fetchForecast(lat, lon)
	if err != nil {
		return nil, ErrServicoIndisponivel
	}

	days := make([]models.ForecastDay, 0, len(data.dailyDates))
	for i := 0; i < len(data.dailyDates) && i < len(data.dailyMin) && i < len(data.dailyMax); i++ {
		days = append(days, models.ForecastDay{
			Data: data.dailyDates[i],
			Min:  data.dailyMin[i],
			Max:  data.dailyMax[i],
		})
	}
	return days, nil
}

type forecastData struct {
	currentTemp float64
	tempMin     float64
	tempMax     float64
	weatherCode int
	humidity    int
	windSpeed   float64
	dailyDates  []string
	dailyMin    []float64
	dailyMax    []float64
}

func fetchForecast(lat, lon float64) (*forecastData, error) {
	endpoint := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&timezone=auto&current_weather=true&hourly=relativehumidity_2m,windspeed_10m&daily=temperature_2m_max,temperature_2m_min&forecast_days=7", lat, lon)
	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, ErrServicoIndisponivel
	}

	client := &http.Client{Timeout: 8 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, ErrServicoIndisponivel
	}
	defer response.Body.Close()

	if response.StatusCode >= 500 {
		return nil, ErrServicoIndisponivel
	}

	var payload struct {
		CurrentWeather struct {
			Temperature float64 `json:"temperature"`
			WeatherCode int     `json:"weathercode"`
			WindSpeed   float64 `json:"windspeed"`
		} `json:"current_weather"`
		Hourly struct {
			Time              []string  `json:"time"`
			RelativeHumidity2 []int     `json:"relativehumidity_2m"`
			WindSpeed10M      []float64 `json:"windspeed_10m"`
		} `json:"hourly"`
		Daily struct {
			Time             []string  `json:"time"`
			Temperature2MMax []float64 `json:"temperature_2m_max"`
			Temperature2MMin []float64 `json:"temperature_2m_min"`
		} `json:"daily"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, ErrServicoIndisponivel
	}

	result := &forecastData{
		currentTemp: payload.CurrentWeather.Temperature,
		tempMin:     firstFloat(payload.Daily.Temperature2MMin),
		tempMax:     firstFloat(payload.Daily.Temperature2MMax),
		weatherCode: payload.CurrentWeather.WeatherCode,
		windSpeed:   payload.CurrentWeather.WindSpeed,
		dailyDates:  payload.Daily.Time,
		dailyMin:    payload.Daily.Temperature2MMin,
		dailyMax:    payload.Daily.Temperature2MMax,
	}

	if len(payload.Hourly.Time) > 0 && len(payload.Hourly.RelativeHumidity2) > 0 {
		result.humidity = payload.Hourly.RelativeHumidity2[len(payload.Hourly.RelativeHumidity2)-1]
	}
	if len(payload.Hourly.WindSpeed10M) > 0 && result.windSpeed == 0 {
		result.windSpeed = payload.Hourly.WindSpeed10M[len(payload.Hourly.WindSpeed10M)-1]
	}

	return result, nil
}

func firstFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	return values[0]
}

func weatherCodeToText(code int) string {
	switch code {
	case 0:
		return "Ensolarado"
	case 1, 2:
		return "Parcialmente Nublado"
	case 3:
		return "Nublado"
	case 45, 48:
		return "Neblina"
	case 51, 53, 55, 56, 57:
		return "Garoa"
	case 61, 63, 65, 80, 81, 82:
		return "Chuva"
	case 71, 73, 75, 77, 85, 86:
		return "Neve"
	default:
		return "Tempo instável"
	}
}

func sampleWeatherForCity(name string) *models.WeatherData {
	key := NormalizeLocal(name)
	switch key {
	case "fortaleza":
		return &models.WeatherData{TempAtual: 29.5, TempMin: 24.0, TempMax: 32.0, Condicao: "Ensolarado", Umidade: 65, VentoKmh: 18.0}
	case "caucaia":
		return &models.WeatherData{TempAtual: 28.2, TempMin: 24.1, TempMax: 31.2, Condicao: "Parcialmente Nublado", Umidade: 68, VentoKmh: 16.0}
	case "sobral":
		return &models.WeatherData{TempAtual: 30.1, TempMin: 23.8, TempMax: 33.0, Condicao: "Ensolarado", Umidade: 52, VentoKmh: 14.0}
	default:
		return &models.WeatherData{TempAtual: 27.0, TempMin: 22.0, TempMax: 31.0, Condicao: "Parcialmente Nublado", Umidade: 60, VentoKmh: 12.0}
	}
}

func sampleForecastForCity(name string) []models.ForecastDay {
	base := sampleWeatherForCity(name)
	start := time.Now().UTC()
	days := make([]models.ForecastDay, 0, 7)
	for i := 0; i < 7; i++ {
		date := start.AddDate(0, 0, i).Format("2006-01-02")
		shift := float64(i) * 0.4
		days = append(days, models.ForecastDay{
			Data: date,
			Min:  round(base.TempMin - shift),
			Max:  round(base.TempMax + shift),
		})
	}
	return days
}

func round(value float64) float64 {
	parsed, _ := strconv.ParseFloat(fmt.Sprintf("%.1f", value), 64)
	return parsed
}

func sortForecastDays(days []models.ForecastDay) {
	sort.Slice(days, func(i, j int) bool { return days[i].Data < days[j].Data })
}
