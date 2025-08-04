package cmds

import (
	"go-hurobot/qbot"
)

func cmd_rps(c *qbot.Client, msg *qbot.Message, args *ArgsList) {
	c.SendMsg(msg, qbot.CQRps())
}
