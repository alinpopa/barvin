package slack

import (
	"encoding/json"
	"fmt"
	"github.com/alinpopa/barvin/data"
	"golang.org/x/net/websocket"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"
)

func startRtm(origin string) data.RtmResponse {
	resp, err := http.Get(origin)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var data data.RtmResponse
	json.NewDecoder(resp.Body).Decode(&data)
	return data
}

func connectWs(url string, origin string) (*websocket.Conn, error) {
	return websocket.Dial(url, "", origin)
}

func currentIpMessage(prefix string) data.WsMessage {
	ipResp, err := http.Get("https://api.ipify.org?format=json")
	if err != nil {
		fmt.Println("Error", err)
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

func replyMessage(ws *websocket.Conn, event data.WsEvent, msg string) error {
	return websocket.JSON.Send(ws, &data.WsEvent{
		Id:      event.Id,
		Type:    "message",
		Channel: event.Channel,
		Text:    msg,
		User:    event.User,
	})
}

func sendPrvMessage(to string, msg string, token string) error {
	rsp, err := http.PostForm("https://slack.com/api/chat.postMessage?token="+token, url.Values{"channel": {to}, "as_user": {"true"}, "text": {msg}})
	fmt.Println("Got resp:", rsp)
	fmt.Println("Got err:", err)
	return err
}

func restart(msg string, err error, c chan<- string) {
	var m string
	if err != nil {
		m = fmt.Sprintf("%s[%s]", msg, err)
	} else {
		m = msg
	}
	fmt.Println(m)
	go func() {
		c <- m
	}()
}

func SlackHandler(initMessage string, restartChannel chan<- string, userId string, token string) {
	origin := "https://slack.com/api/rtm.start?token=" + token
	rtm := startRtm(origin)
	fmt.Println(rtm.Url)
	ws, err := connectWs(rtm.Url, origin)
	for err != nil {
		time.Sleep(3 * time.Second)
		rtm := startRtm(origin)
		ws, err = connectWs(rtm.Url, origin)
	}
	sendPrvMessage(userId, currentIpMessage(initMessage).Msg, token)
	for {
		var msg string
		var event data.WsEvent
		err := websocket.Message.Receive(ws, &msg)
		if err != nil {
			restart("Error while receiving message", err, restartChannel)
			break
		}
		unmarshallErr := json.Unmarshal([]byte(msg), &event)
		if unmarshallErr != nil {
			fmt.Println("Error while unmarshaling message:", msg)
			restart("Error unmarshalling message", unmarshallErr, restartChannel)
			break
		}
		fmt.Println("Raw message:", msg)
		fmt.Printf("Got event %+v\n", event)
		if strings.ToLower(event.Text) == "ip" && event.User == userId {
			replyMessage(ws, event, currentIpMessage("").Msg)
		}
		runtime.Gosched()
	}
}
