package cmds

import (
	"go-hurobot/qbot"
	"strconv"

	"github.com/buger/jsonparser"
)

func cmd_specialtitle(c *qbot.Client, args []string, raw *qbot.Message) {
	if len(args) == 1 {
		c.SendReplyMsg(raw, "Usage: specialtitle <specialtitle>")
	} else if len(raw.Message) > 1 && raw.Message[1].Type != "at" {
		c.SendReplyMsg(raw, "群头衔一定是一个文本！")
	} else if length := len([]byte(args[1])); length > 18 {
		c.SendReplyMsg(raw, "头衔长度不允许超过 18 字节，当前 "+strconv.FormatInt(int64(length), 10)+" 字节")
	} else {
		if len(raw.Message) > 1 {
			rawJson := raw.Message[1].Data
			qqstr, err := jsonparser.GetString([]byte(rawJson), "qq")
			if err != nil {
				c.SendReplyMsg(raw, err.Error())
				return
			}
			qq, err := strconv.Atoi(qqstr)
			if err != nil {
				c.SendReplyMsg(raw, err.Error())
				return
			}
			c.SetGroupSpecialTitle(raw.GroupID, uint64(qq), args[1])
			return
		}
		c.SetGroupSpecialTitle(raw.GroupID, raw.Sender.UserID, args[1])
	}
}
