package external

import (
	"encoding/json"
	"fmt"
	bdata "github.com/alinpopa/barvin/data"
	"github.com/alinpopa/barvin/handlers/slack/data"
	"net/http"
)

func GetHomeInWeather() data.WsMessage {
	weatherResp, err := http.Get("https://agent.electricimp.com/ABsCu4F5e-PU/json")
	if err != nil {
		log.Errorf("Error while fetching the weather: %s", err)
		return data.WsMessage{Msg: fmt.Sprintf("*Error*: %s", err)}
	}
	defer weatherResp.Body.Close()
	var weatherInfo bdata.WeatherInfo
	json.NewDecoder(weatherResp.Body).Decode(&weatherInfo)
	msgFormat := "" +
		"*In Weather*\n" +
		"  temperature: %.2f\n" +
		"  pressure: %.2f\n" +
		"  last pressure: %.2f\n" +
		"  is day: %t\n" +
		"  humidity: %.2f\n" +
		"  lux: %.3f\n" +
		"  date: %s"
	return data.WsMessage{Msg: fmt.Sprintf(msgFormat,
		weatherInfo.Temp,
		weatherInfo.Pressure,
		weatherInfo.LastPressure,
		weatherInfo.Day,
		weatherInfo.Humidity,
		weatherInfo.Lux,
		weatherInfo.Date)}
}

func GetHomeOutWeather() data.WsMessage {
	weatherResp, err := http.Get("https://agent.electricimp.com/nBWDNt16ALCQ/json")
	if err != nil {
		log.Errorf("Error while fetching the weather: %s", err)
		return data.WsMessage{Msg: fmt.Sprintf("*Error*: %s", err)}
	}
	defer weatherResp.Body.Close()
	var weatherInfo bdata.WeatherInfo
	json.NewDecoder(weatherResp.Body).Decode(&weatherInfo)
	msgFormat := "" +
		"*Out Weather*\n" +
		"  temperature: %.2f\n" +
		"  pressure: %.2f\n" +
		"  last pressure: %.2f\n" +
		"  humidity: %.2f\n" +
		"  date: %s"
	return data.WsMessage{Msg: fmt.Sprintf(msgFormat,
		weatherInfo.Temp,
		weatherInfo.Pressure,
		weatherInfo.LastPressure,
		weatherInfo.Humidity,
		weatherInfo.Date)}
}
