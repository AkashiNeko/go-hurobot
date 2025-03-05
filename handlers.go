package main

import (
	"go-hurobot/cmds"
	"go-hurobot/qbot"
	"strings"

	"github.com/google/shlex"
)

func getCommandName(s string) string {
	const maxLength = 20
	singleCmd := false
	if len(s) > maxLength {
		s = s[:maxLength]
	} else {
		singleCmd = true
	}

	if index := strings.Index(s, " "); index != -1 {
		return s[:index]
	} else if singleCmd {
		return s
	}
	return ""
}

func onMessage(c *qbot.Client, msg *qbot.Message) {
	if handler := cmds.FindCommand(getCommandName(msg.RawMessage)); handler != nil {
		if args, err := shlex.Split(msg.RawMessage); err == nil {
			handler(c, args, msg)
		}
		return
	}

	// for i, arg := range args {
	// 	res := fmt.Sprintf("args[%d]: %q", i, arg)
	// 	c.SendPrivateMsg(msg.Sender.UserID, res, false)
	// }
}
