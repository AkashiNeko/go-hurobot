package main

import "go-hurobot/qbot"

func onGroupMessage(c *qbot.Client, msg qbot.GroupMessage) {
	if msg.RawMessage == "hello" {
		c.SendGroupMsg(msg.GroupID, "world", false)
	}
}

func onPrivateMessage(c *qbot.Client, msg qbot.PrivateMessage) {
	if msg.RawMessage == "hello" {
		c.SendPrivateMsg(msg.Sender.UserID, "world", false)
	}
}
