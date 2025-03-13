package cmds

import (
	"go-hurobot/qbot"
	"strconv"
)

func cmd_specialtitle(c *qbot.Client, args []string, msg *qbot.Message) {
	if len(args) == 1 {
		c.SendReplyMsg(msg, "Usage: specialtitle <specialtitle>")
	} else if len(msg.Array) > 1 && msg.Array[1].Type != qbot.At {
		c.SendReplyMsg(msg, "群头衔一定是一个文本！")
	} else if length := len([]byte(args[1])); length > 18 {
		c.SendReplyMsg(msg, "头衔长度不允许超过 18 字节，当前 "+strconv.FormatInt(int64(length), 10)+" 字节")
	} else {
		if len(msg.Array) > 1 {
			id := str2uin64(msg.Array[1].Content)
			c.SetGroupSpecialTitle(msg.GroupID, id, args[1])
			return
		}
		c.SetGroupSpecialTitle(msg.GroupID, msg.UserID, args[1])
	}
}
