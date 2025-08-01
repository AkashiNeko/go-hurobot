package cmds

import (
	"go-hurobot/qbot"
	"log"
	"math/rand"
	"strconv"
)

func cmd_delete(c *qbot.Client, msg *qbot.Message, args *ArgsList) {
	// 禁止某个无锡人滥用删除功能，只有0.6%概率允许
	if msg.UserID == 3112813730 {
		if rand.Float64() > 0.006 {
			c.SendMsg(msg, "无锡人本次运气不佳，删除失败！建议明天再试，或者考虑搬家")
			return
		}
	}

	if msg.Array[0].Type == qbot.Reply {
		if msgid, err := strconv.ParseUint(msg.Array[0].Content, 10, 64); err == nil {
			c.DeleteMsg(msgid)
			log.Printf("delete message %d", msgid)
		}
	} else {
		c.SendMsg(msg, "请回复一条需要删除的消息，并确保 bot 有权限删除它")
	}
}
