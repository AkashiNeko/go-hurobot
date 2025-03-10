package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-hurobot/config"
	"go-hurobot/qbot"
)

func main() {
	bot := qbot.NewClient(&qbot.Config{
		Address:      config.NapcatWSURL,
		Reconnect:    3 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})
	defer bot.Close()

	bot.HandleMessage(onMessage)

	// time.Sleep(1 * time.Second)
	// _, err := bot.SendPrivateMsg(config.AdminId, "Hello master!", false)
	// if err != nil {
	// 	log.Printf("send message failed: %v", err)
	// }

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	stopSignal := <-stop
	fmt.Println("Bot is shutting down:", stopSignal.String())
}
