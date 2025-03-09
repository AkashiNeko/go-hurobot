package main

import (
	"go-hurobot/qbot"
	"strings"
)

func customReply(c *qbot.Client, msg *qbot.Message) {
	// 2025-03-08 晚上，让 bot 在某 mc 群发电加的
	if msg.GroupID == 158045531 {
		if strings.Contains(msg.RawMessage, "厉厉厉害害害") {
			return
		} else if strings.Contains(msg.RawMessage, "厉厉害害") {
			c.SendGroupMsg(msg.GroupID, strings.Replace(msg.RawMessage, "厉厉害害", "可可爱爱", -1), false)
		} else if strings.Contains(msg.RawMessage, "厉害") {
			c.SendGroupMsg(msg.GroupID, strings.Replace(msg.RawMessage, "厉害", "可爱", -1), false)
		}
	}
	// 2025-03-08 ↑
}
