package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-hurobot/qbot"
)

const MASTER_ID uint64 = 1006554341

func main() {
	bot := qbot.NewClient(&qbot.Config{
		Address:      "ws://localhost:3001",
		Reconnect:    3 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})
	defer bot.Close()

	bot.HandleGroupMessage(onGroupMessage)
	bot.HandlePrivateMessage(onPrivateMessage)

	time.Sleep(1 * time.Second)
	_, err := bot.SendPrivateMsg(MASTER_ID, "Hello master!", false)
	if err != nil {
		log.Printf("send message failed: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	stopSignal := <-stop
	fmt.Println("Bot is shutting down:", stopSignal.String())
}
