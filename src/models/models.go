package models

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

type ForecastDay struct {
	Data string  `json:"data"`
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
