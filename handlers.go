package main

import (
	"go-hurobot/cmds"
	"go-hurobot/qbot"
	"strings"

	"github.com/google/shlex"
)

func getCommandName(s string) string {
	sliced := false
	if len(s) > cmds.MaxCommandLength {
		s = s[:cmds.MaxCommandLength]
		sliced = true
	}

	i := strings.IndexAny(s, " \n")
	if i != -1 {
		return s[:i]
	}
	if sliced {
		return ""
	}
	return s
}

func onMessage(c *qbot.Client, msg *qbot.Message) {
	if handler := cmds.FindCommand(getCommandName(msg.RawMessage)); handler != nil {
		if args, err := shlex.Split(msg.RawMessage); err == nil {
			handler(c, args, msg)
		}
		return
	}
	customReply(c, msg)
}
