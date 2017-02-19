package external

import (
	"fmt"
	"github.com/alinpopa/barvin/handlers/slack/data"
	"io/ioutil"
	"net/http"
	"net/url"
)

func GetMacInfo(mac string) data.WsMessage {
	macResp, err1 := http.Get(fmt.Sprintf("http://api.macvendors.com/%s", url.QueryEscape(mac)))
	if err1 != nil {
		log.Errorf("Error while fetching the mac: %s", err1)
		return data.WsMessage{Msg: fmt.Sprintf("```Error: %s```", err1)}
	}
	defer macResp.Body.Close()
	textBody, err2 := ioutil.ReadAll(macResp.Body)
	if err2 != nil {
		log.Errorf("Error while fetching the mac: %s", err2)
		return data.WsMessage{Msg: fmt.Sprintf("```Error: %s```", err2)}
	}
	msgFormat := "" +
		"```\n" +
		"mac vendor: %s\n" +
		"```"
	return data.WsMessage{Msg: fmt.Sprintf(msgFormat, textBody)}
}
