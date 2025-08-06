package main

import (
	"go-hurobot/cmds"
	"go-hurobot/config"
	"go-hurobot/llm"
	"go-hurobot/mc"
	"go-hurobot/qbot"
)

func messageHandler(c *qbot.Client, msg *qbot.Message) {
	if msg.UserID != config.BotID {
		isCommand := cmds.HandleCommand(c, msg)
		defer qbot.SaveDatabase(msg, isCommand)

		// Forward non-command messages to Minecraft if RCON is enabled
		if !isCommand {
			mc.ForwardMessageToMC(c, msg)
		}

		if isCommand {
			return
		}
		if llm.LLMMsgHandle(c, msg) {
			return
		}
		if cmds.CheckUserEvents(c, msg) {
			return
		}
		customReply(c, msg)
	}
}
