package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go-hurobot/cmds"
	"go-hurobot/qbot"
)

func main() {
	bot := qbot.NewClient()
	defer bot.Close()

	bot.HandleMessage(func(c *qbot.Client, msg *qbot.Message) {
		cmds.HandleCommand(c, msg)
		customReply(c, msg)
	})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	stopSignal := <-stop
	fmt.Println("shutting down:", stopSignal.String())
}
