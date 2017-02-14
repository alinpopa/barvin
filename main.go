package main

import (
	"flag"
	//"fmt"
	"github.com/alinpopa/barvin/handlers/slack"
	"github.com/op/go-logging"
	"os"
	"sync"
	"time"
)

func main() {
	userID := flag.String("userid", "", "The privileged slack userid.")
	token := flag.String("token", "", "Slack token to connect.")
	flag.Parse()

	var format = logging.MustStringFormatter(
		`%{color}%{time:2006-01-02T15:04:05.000} %{shortfunc} >>> %{level} %{id:03x} %{message}%{color:reset}`,
	)
	loggingBackend := logging.NewLogBackend(os.Stderr, "", 0)
	backend2Formatter := logging.NewBackendFormatter(loggingBackend, format)
	loggingBackendLeveled := logging.AddModuleLevel(loggingBackend)
	loggingBackendLeveled.SetLevel(logging.DEBUG, "")
	logging.SetBackend(backend2Formatter)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		restartChannel := make(chan string)
		go func() {
			restartChannel <- "Initial start"
		}()
		for {
			msg := <-restartChannel
			_ = slack.SlackHandler(msg, restartChannel, *userID, *token)
			time.Sleep(35 * time.Second)
		}
	}()
	wg.Wait()
}
