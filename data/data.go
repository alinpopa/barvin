package data

type IpInfo struct {
	Ip string `json:"ip"`
}

type WeatherInfo struct {
	Temp         float32 `json:"temp"`
	Pressure     float32 `json:"pressure"`
	Day          bool    `json:"day"`
	Humidity     float32 `json:"humid"`
	Lux          float32 `json:"lux"`
	LastPressure float32 `json:"lastPressure"`
	Date         string  `json:"date"`
}
