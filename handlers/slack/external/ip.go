package external

import (
	"encoding/json"
	"fmt"
	bdata "github.com/alinpopa/barvin/data"
	"github.com/alinpopa/barvin/handlers/slack/data"
	"github.com/op/go-logging"
	"net/http"
)

var log = logging.MustGetLogger("main-logger")

func CurrentIpMessage(prefix string) data.WsMessage {
	ipResp, err := http.Get("https://api.ipify.org?format=json")
	if err != nil {
		log.Errorf("Error while fetching the IP: %s", err)
		return data.WsMessage{Msg: fmt.Sprintf("*Error*: %s", err)}
	}
	defer ipResp.Body.Close()
	var ipInfo bdata.IpInfo
	json.NewDecoder(ipResp.Body).Decode(&ipInfo)
	if len(prefix) > 0 {
		msgFormat := "" +
			"*%s*\n" +
			"  ip: %s"
		return data.WsMessage{Msg: fmt.Sprintf(msgFormat, prefix, ipInfo.Ip)}
	}
	msgFormat := "ip: %s"
	return data.WsMessage{Msg: fmt.Sprintf(msgFormat, ipInfo.Ip)}
}
