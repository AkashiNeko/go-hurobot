package cmds

import (
	"go-hurobot/qbot"
	"strings"
)

func cmd_echo(c *qbot.Client, args []string, raw *qbot.Message) {
	c.SendReplyMsg(raw, strings.Join(args[1:], " "))
}
