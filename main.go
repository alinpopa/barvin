package main

import (
	"flag"
	"github.com/alinpopa/barvin/handlers/slack"
	"sync"
)

func main() {
	userId := flag.String("userid", "", "The privileged slack userid.")
	token := flag.String("token", "", "Slack token to connect.")
	flag.Parse()

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
			go slack.SlackHandler(msg, restartChannel, *userId, *token)
		}
	}()
	wg.Wait()
}
