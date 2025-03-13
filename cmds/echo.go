package cmds

import (
	"go-hurobot/qbot"
	"strings"
)

func cmd_echo(c *qbot.Client, msg *qbot.Message, args *ArgsList) {
	c.SendReplyMsg(msg, strings.Trim(msg.Raw[4:], " \n"))
}
