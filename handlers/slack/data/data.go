package data

import (
	"github.com/gorilla/websocket"
	"github.com/op/go-logging"
	"time"
)

var log = logging.MustGetLogger("main-logger")

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

type Checker struct {
	StopPinger      chan bool
	StopChecker     chan bool
	Alive           chan bool
	StopHandlerChan chan bool
	Ws              *websocket.Conn
}

type Context struct {
	Checker *Checker
	Ws      *websocket.Conn
	User    string
}

func (ctx *Context) RunChecker() {
	go func() {
		for {
			select {
			case <-ctx.Checker.StopPinger:
				return
			case <-time.After(10 * time.Second):
				ctx.Ws.WriteJSON(&WsEvent{
					Id:   911,
					Type: "ping",
				})
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Checker.StopChecker:
				return
			case <-ctx.Checker.Alive:
			case <-time.After(30 * time.Second):
				log.Error("Connection timeout after 30 seconds; trying to restart...")
				ctx.Checker.StopHandlerChan <- true
				return
			}
		}
	}()
}

func (ctx *Context) StopChecker() {
	ctx.Checker.StopPinger <- true
	ctx.Checker.StopChecker <- true
}

func (ctx *Context) Stop() {
	ctx.StopChecker()
	ctx.Ws.Close()
}
