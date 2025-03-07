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

	if i := strings.Index(s, " "); i != -1 {
		return s[:i]
	} else if i := strings.Index(s, "\n"); i != -1 {
		return s[:i]
	} else if singleCmd {
		return s
	} else {
		return ""
	}
}

func onMessage(c *qbot.Client, msg *qbot.Message) {
	if handler := cmds.FindCommand(getCommandName(msg.RawMessage)); handler != nil {
		if args, err := shlex.Split(msg.RawMessage); err == nil {
			handler(c, args, msg)
		}
		return
	}
}
