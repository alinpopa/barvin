package external

import (
	"encoding/json"
	"fmt"
	bdata "github.com/alinpopa/barvin/data"
	"github.com/alinpopa/barvin/handlers/slack/data"
	"net/http"
)

func GetHomeWeather() data.WsMessage {
	weatherResp, err := http.Get("https://agent.electricimp.com/ABsCu4F5e-PU/json")
	if err != nil {
		log.Errorf("Error while fetching the weather: %s", err)
		return data.WsMessage{Msg: fmt.Sprintf("Error: %s", err)}
	}
	defer weatherResp.Body.Close()
	var weatherInfo bdata.WeatherInfo
	json.NewDecoder(weatherResp.Body).Decode(&weatherInfo)
	msgFormat := "" +
		"```\n" +
		"Weather\n" +
		"temperature: %f\n" +
		"pressure: %f\n" +
		"last pressure: %f\n" +
		"is day: %t\n" +
		"humidity: %f\n" +
		"lux: %f\n" +
		"date: %s\n" +
		"```"
	return data.WsMessage{Msg: fmt.Sprintf(msgFormat,
		weatherInfo.Temp,
		weatherInfo.Pressure,
		weatherInfo.LastPressure,
		weatherInfo.Day,
		weatherInfo.Humidity,
		weatherInfo.Lux,
		weatherInfo.Date)}
}
