package main

import (
	"go-hurobot/cmds"
	"go-hurobot/config"
	"go-hurobot/qbot"
)

func messageHandler(c *qbot.Client, msg *qbot.Message) {
	if msg.UserID != config.BotID {
		isCommand := cmds.HandleCommand(c, msg)
		go qbot.SaveDatabase(msg, isCommand)
		customReply(c, msg)
	}
}
