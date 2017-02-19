package commands

import (
	"github.com/alinpopa/barvin/handlers/slack/data"
	"github.com/alinpopa/barvin/handlers/slack/external"
	"github.com/op/go-logging"
	"strings"
)

var log = logging.MustGetLogger("main-logger")

type Command interface {
	Run() error
}

type NoopCommand struct {
}

type PongCommand struct {
	Ctx *data.Context
}

type IpCommand struct {
	Ctx   *data.Context
	Event *data.WsEvent
}

type WeatherCommand struct {
	Ctx   *data.Context
	Event *data.WsEvent
}

type GetMacCommand struct {
	Ctx   *data.Context
	Event *data.WsEvent
	Mac   string
}

func (cmd *NoopCommand) Run() error {
	return nil
}

func (cmd *PongCommand) Run() error {
	cmd.Ctx.Checker.Alive <- true
	return nil
}

func (cmd *IpCommand) Run() error {
	msg := external.CurrentIpMessage("").Msg
	return cmd.Ctx.Ws.WriteJSON(&data.WsEvent{
		Id:      cmd.Event.Id,
		Type:    "message",
		Channel: cmd.Event.Channel,
		Text:    msg,
		User:    cmd.Event.User,
	})
}

func (cmd *WeatherCommand) Run() error {
	msg := external.GetHomeWeather().Msg
	return cmd.Ctx.Ws.WriteJSON(&data.WsEvent{
		Id:      cmd.Event.Id,
		Type:    "message",
		Channel: cmd.Event.Channel,
		Text:    msg,
		User:    cmd.Event.User,
	})
}

func (cmd *GetMacCommand) Run() error {
	msg := external.GetMacInfo(cmd.Mac).Msg
	return cmd.Ctx.Ws.WriteJSON(&data.WsEvent{
		Id:      cmd.Event.Id,
		Type:    "message",
		Channel: cmd.Event.Channel,
		Text:    msg,
		User:    cmd.Event.User,
	})
}

func EventToCommand(event *data.WsEvent, ctx *data.Context) Command {
	if strings.ToLower(event.Type) == "pong" {
		return &PongCommand{
			Ctx: ctx,
		}
	}
	if event.User != ctx.User {
		return &NoopCommand{}
	}
	if strings.ToLower(event.Text) == "ip" {
		return &IpCommand{
			Ctx:   ctx,
			Event: event,
		}
	}
	if strings.ToLower(event.Text) == "weather" {
		return &WeatherCommand{
			Ctx:   ctx,
			Event: event,
		}
	}
	var maybeComplexCmd = strings.Split(strings.ToLower(event.Text), " ")
	if len(maybeComplexCmd) != 2 {
		return &NoopCommand{}
	}
	msg := maybeComplexCmd[0]
	arg := maybeComplexCmd[1]
	if msg == "mac" {
		return &GetMacCommand{
			Ctx:   ctx,
			Event: event,
			Mac:   arg,
		}
	}
	return &NoopCommand{}
}
