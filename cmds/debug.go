package cmds

import (
	"go-hurobot/config"
	"go-hurobot/qbot"
	"strings"
)

func cmd_debug(c *qbot.Client, msg *qbot.Message, args *ArgsList) {
	if msg.UserID == config.MasterID {
		c.SendMsg(msg, decodeSpecialChars(strings.Trim(msg.Raw[6:], " \n")))
	}

}
