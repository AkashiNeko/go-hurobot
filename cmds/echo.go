package cmds

import (
	"go-hurobot/qbot"
	"strings"
)

func cmd_echo(c *qbot.Client, args []string, msg *qbot.Message) {
	c.SendReplyMsg(msg, strings.Trim(msg.Raw[4:], " \n"))
}
