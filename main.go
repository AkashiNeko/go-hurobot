package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go-hurobot/qbot"
)

func main() {
	bot := qbot.NewClient()
	defer bot.Close()

	bot.HandleMessage(messageHandler)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	stopSignal := <-stop
	fmt.Println("shutting down:", stopSignal.String())
}
