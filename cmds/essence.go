package cmds

import (
	"go-hurobot/config"
	"go-hurobot/qbot"
	"slices"
	"strconv"
)

func cmd_essence(c *qbot.Client, raw *qbot.Message, args *ArgsList) {
	if !slices.Contains(config.BotOwnerGroupIDs, raw.GroupID) {
		return
	}
	help := "请回复一条消息，再使用 essence [set|delete]"
	if raw.Array[0].Type != qbot.Reply {
		c.SendMsg(raw, help)
		return
	}
	msgID, err := strconv.ParseUint(raw.Array[0].Content, 10, 64)
	if err != nil {
		return
	}
	if args.Size == 2 {
		if args.Contents[1] == "delete" {
			c.DeleteGroupEssence(msgID)
		} else if args.Contents[1] == "set" {
			c.SetGroupEssence(msgID)
		} else {
			c.SendMsg(raw, help)
		}
	} else if args.Size == 1 {
		c.SetGroupEssence(msgID)
	} else {
		c.SendMsg(raw, help)
	}
}
