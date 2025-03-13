package main

import (
	"go-hurobot/cmds"
	"go-hurobot/qbot"
	"strings"

	"github.com/google/shlex"
)

func getCommandName(s string) string {
	sliced := false
	if len(s) > cmds.MaxCommandLength+1 {
		s = s[:cmds.MaxCommandLength+1]
		sliced = true
	}
	if i := strings.IndexAny(s, " \n"); i != -1 {
		return s[:i]
	}
	if sliced {
		return ""
	}
	return s
}

func onMessage(c *qbot.Client, msg *qbot.Message) {
	if handler := cmds.FindCommand(getCommandName(msg.Raw)); handler != nil {
		if args, err := shlex.Split(msg.Raw); err == nil {
			handler(c, args, msg)
		}
		return
	}
	customReply(c, msg)
}
