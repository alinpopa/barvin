package data

type WsError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type WsEvent struct {
	Id      int     `json:"id"`
	Type    string  `json:"type"`
	Channel string  `json:"channel"`
	Text    string  `json:"text"`
	Ok      bool    `json:"ok"`
	ReplyTo int     `json:"reply_to"`
	Ts      string  `json:"ts"`
	Error   WsError `json:"error"`
	Url     string  `json:"url"`
	User    string  `json:"user"`
}

type IpInfo struct {
	Ip string `json:"ip"`
}

type WeatherInfo struct {
	Temp     float32 `json:"temp"`
	Pressure float32 `json:"pressure"`
	Day      bool    `json:"day"`
	Humidity float32 `json:"humid"`
	Lux      float32 `json:"lux"`
}

type RtmResponse struct {
	Url string `json:"url"`
}

type WsMessage struct {
	Msg string
}

type SlackUser struct {
	User    string
	Channel string
}
