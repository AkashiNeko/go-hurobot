package cmds

import (
	"go-hurobot/qbot"
	"strings"
)

func echo(c *qbot.Client, args []string, raw *qbot.Message) {
	c.SendReplyMsg(raw, strings.Join(args[1:], " "))
}
