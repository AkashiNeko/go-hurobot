package cmds

import "go-hurobot/qbot"

var MaxCommandLength int = 0

var cmdMap map[string]func(*qbot.Client, []string, *qbot.Message)

func init() {
	cmdMap = map[string]func(*qbot.Client, []string, *qbot.Message){
		"echo":         cmd_echo,
		"rawmsg":       cmd_rawmsg,
		"grok2":        cmd_grok2,
		"specialtitle": cmd_specialtitle,
		"sh":           cmd_sh,
	}

	for key := range cmdMap {
		if len(key) > MaxCommandLength {
			MaxCommandLength = len(key)
		}
	}
}

func FindCommand(cmd string) func(*qbot.Client, []string, *qbot.Message) {
	if cmd == "" {
		return nil
	}
	return cmdMap[cmd]
}
