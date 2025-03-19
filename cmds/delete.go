package cmds

import (
	"go-hurobot/qbot"
	"log"
	"strconv"
)

func cmd_delete(c *qbot.Client, msg *qbot.Message, args *ArgsList) {
	if msg.Array[0].Type == qbot.Reply {
		if msgid, err := strconv.ParseUint(msg.Array[0].Content, 10, 64); err == nil {
			c.DeleteMsg(msgid)
			log.Printf("delete message %d", msgid)
		}
	} else {
		c.SendMsg(msg, "请回复一条需要删除的消息，并确保 bot 有权限删除它")
	}
}
