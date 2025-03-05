package cmds

import (
	"go-hurobot/qbot"
	"strings"
)

func Echo(c *qbot.Client, args []string, raw *qbot.Message) {
	c.SendReplyMsg(raw, strings.Join(args[1:], " "))
}
