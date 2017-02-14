package slack

import (
	"encoding/json"
	"fmt"
	"github.com/alinpopa/barvin/data"
	"github.com/gorilla/websocket"
	"github.com/op/go-logging"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var log = logging.MustGetLogger("main-logger")

func startRtm(origin string) data.RtmResponse {
	resp, err := http.Get(origin)
	if err != nil {
		return data.RtmResponse{}
	}
	defer resp.Body.Close()
	var data data.RtmResponse
	json.NewDecoder(resp.Body).Decode(&data)
	return data
}

func connectWs(url string, origin string) (*websocket.Conn, error) {
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	return c, err
}

func currentIpMessage(prefix string) data.WsMessage {
	ipResp, err := http.Get("https://api.ipify.org?format=json")
	if err != nil {
		log.Errorf("Error while fetching the IP: %s", err)
		return data.WsMessage{Msg: fmt.Sprintf("Error: %s", err)}
	}
	defer ipResp.Body.Close()
	var ipInfo data.IpInfo
	json.NewDecoder(ipResp.Body).Decode(&ipInfo)
	if len(prefix) > 0 {
		return data.WsMessage{Msg: prefix + ":" + ipInfo.Ip}
	}
	return data.WsMessage{Msg: ipInfo.Ip}
}

func getHomeWeather() data.WsMessage {
	weatherResp, err := http.Get("https://agent.electricimp.com/ABsCu4F5e-PU/json")
	if err != nil {
		log.Errorf("Error while fetching the weather: %s", err)
		return data.WsMessage{Msg: fmt.Sprintf("Error: %s", err)}
	}
	defer weatherResp.Body.Close()
	var weatherInfo data.WeatherInfo
	json.NewDecoder(weatherResp.Body).Decode(&weatherInfo)
	return data.WsMessage{Msg: fmt.Sprintf("Weather\ntemperature:%f\npressure:%f\nlastPressure:%f\nis day:%t\nhumidity:%f\nlux:%f",
		weatherInfo.Temp,
		weatherInfo.Pressure,
		weatherInfo.LastPressure,
		weatherInfo.Day,
		weatherInfo.Humidity,
		weatherInfo.Lux)}
}

func getMac(mac string) data.WsMessage {
	macResp, err1 := http.Get(fmt.Sprintf("http://api.macvendors.com/%s", url.QueryEscape(mac)))
	if err1 != nil {
		log.Errorf("Error while fetching the mac: %s", err1)
		return data.WsMessage{Msg: fmt.Sprintf("Error: %s", err1)}
	}
	defer macResp.Body.Close()
	textBody, err2 := ioutil.ReadAll(macResp.Body)
	if err2 != nil {
		log.Errorf("Error while fetching the mac: %s", err2)
		return data.WsMessage{Msg: fmt.Sprintf("Error: %s", err2)}
	}
	return data.WsMessage{Msg: fmt.Sprintf("Mac vendor: %s", textBody)}
}

func replyMessage(ws *websocket.Conn, event data.WsEvent, msg string) error {
	return ws.WriteJSON(&data.WsEvent{
		Id:      event.Id,
		Type:    "message",
		Channel: event.Channel,
		Text:    msg,
		User:    event.User,
	})
}

func sendPrvMessage(to string, msg string, token string) error {
	rsp, err := http.PostForm("https://slack.com/api/chat.postMessage?token="+token, url.Values{"channel": {to}, "as_user": {"true"}, "text": {msg}})
	if err != nil {
		log.Errorf("Error while sending private message: %s", err)
	}
	if rsp != nil {
		rsp.Body.Close()
	}
	return err
}

func restart(msg string, err error, c chan<- string) {
	var m string
	if err != nil {
		m = fmt.Sprintf("%s[%s]", msg, err)
	} else {
		m = msg
	}
	log.Noticef("Restart: %s", m)
	go func() {
		c <- m
	}()
}

func parseEventText(eventText string) []string {
	return strings.Split(eventText, " ")
}

func startPinger(ws *websocket.Conn, stopPinger chan bool) {
	go func() {
		for {
			select {
			case <-stopPinger:
				return
			case <-time.After(10 * time.Second):
				ws.WriteJSON(&data.WsEvent{
					Id:   911,
					Type: "ping",
				})
			}
		}
	}()
}

func eventsProducer(ws *websocket.Conn, eventsChan chan<- data.WsEvent, stopHandlerChan chan<- bool) {
	var event data.WsEvent
	_, msg, err := ws.ReadMessage()
	if err != nil {
		log.Errorf("Error while reading message: %s", err)
		stopHandlerChan <- true
		return
	}
	unmarshallErr := json.Unmarshal([]byte(msg), &event)
	if unmarshallErr != nil {
		log.Errorf("Error while unmarshaling message: %s", msg)
		stopHandlerChan <- true
		return
	}
	log.Debugf("Raw message: %s", msg)
	log.Infof("Got event %+v", event)
	eventsChan <- event
}

type checker struct {
	stopPinger      chan bool
	stopChecker     chan bool
	alive           chan bool
	stopHandlerChan chan bool
	ws              *websocket.Conn
}

func (checker *checker) run() {
	startPinger(checker.ws, checker.stopPinger)
	go func() {
		for {
			select {
			case <-checker.stopChecker:
				return
			case <-checker.alive:
			case <-time.After(30 * time.Second):
				log.Error("Connection timeout after 30 seconds; trying to restart...")
				checker.stopHandlerChan <- true
				return
			}
		}
	}()
}

func (checker *checker) stop() {
	checker.stopPinger <- true
	checker.stopChecker <- true
}

func createChecker(ws *websocket.Conn, stopHandlerChan chan bool) *checker {
	return &checker{
		stopPinger:      make(chan bool),
		stopChecker:     make(chan bool),
		alive:           make(chan bool),
		stopHandlerChan: stopHandlerChan,
		ws:              ws,
	}
}

func SlackHandler(initMessage string, restartChan chan<- string, userId string, token string) chan<- bool {
	stopHandlerChan := make(chan bool)
	go func() {
		origin := "https://slack.com/api/rtm.start?token=" + token
		rtm := startRtm(origin)
		log.Debugf("RTM url: %s", rtm.Url)
		ws, err := connectWs(rtm.Url, origin)
		for err != nil {
			log.Infof("Got error; trying to connect to WS: %s", err)
			restart("Got error while trying to connect to WS", err, restartChan)
			return
		}
		checker := createChecker(ws, stopHandlerChan)
		defer ws.Close()
		defer checker.stop()
		checker.run()
		sendPrvMessage(userId, currentIpMessage(initMessage).Msg, token)
		sendPrvMessage(userId, getHomeWeather().Msg, token)
		eventsChan := make(chan data.WsEvent)
		go eventsProducer(ws, eventsChan, stopHandlerChan)
		for {
			select {
			case <-stopHandlerChan:
				restart("Got handler stop signal; restarting handler..", nil, restartChan)
				return
			case event := <-eventsChan:
				if strings.ToLower(event.Type) == "pong" {
					checker.alive <- true
				} else if strings.ToLower(event.Text) == "ip" && event.User == userId {
					replyMessage(ws, event, currentIpMessage("").Msg)
				} else if strings.ToLower(event.Text) == "weather" && event.User == userId {
					replyMessage(ws, event, getHomeWeather().Msg)
				} else if event.User == userId {
					eventText := parseEventText(event.Text)
					if len(eventText) == 2 {
						msg := strings.ToLower(eventText[0])
						if msg == "mac" {
							replyMessage(ws, event, getMac(eventText[1]).Msg)
						}
					}
				}
				go eventsProducer(ws, eventsChan, stopHandlerChan)
			}
		}
	}()
	return stopHandlerChan
}
