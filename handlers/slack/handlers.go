package slack

import (
	"encoding/json"
	"fmt"
	"github.com/alinpopa/barvin/handlers/slack/commands"
	"github.com/alinpopa/barvin/handlers/slack/data"
	"github.com/alinpopa/barvin/handlers/slack/external"
	"github.com/gorilla/websocket"
	"github.com/op/go-logging"
	"net/http"
	"net/url"
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
		event = data.WsEvent{
			Id:   0,
			Type: "DUMMY",
		}
	}
	log.Debugf("Raw message: %s", msg)
	log.Infof("Got event %+v", event)
	eventsChan <- event
}

func startContext(ws *websocket.Conn, stopHandlerChan chan bool, user string) *data.Context {
	ctx := &data.Context{
		Ws: ws,
		Checker: &data.Checker{
			StopPinger:      make(chan bool),
			StopChecker:     make(chan bool),
			Alive:           make(chan bool),
			StopHandlerChan: stopHandlerChan,
		},
		User: user,
	}
	ctx.RunChecker()
	return ctx
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
		context := startContext(ws, stopHandlerChan, userId)
		defer context.Stop()
		sendPrvMessage(userId, external.CurrentIpMessage(initMessage).Msg, token)
		sendPrvMessage(userId, external.GetHomeInWeather().Msg, token)
		sendPrvMessage(userId, external.GetHomeOutWeather().Msg, token)
		eventsChan := make(chan data.WsEvent)
		go eventsProducer(ws, eventsChan, stopHandlerChan)
		for {
			select {
			case <-stopHandlerChan:
				restart("Got handler stop signal; restarting handler..", nil, restartChan)
				return
			case event := <-eventsChan:
				commands.EventToCommand(&event, context).Run()
				go eventsProducer(ws, eventsChan, stopHandlerChan)
			}
		}
	}()
	return stopHandlerChan
}
