package cmds

import "go-hurobot/qbot"

var cmdMap map[string]func(*qbot.Client, []string, *qbot.Message)

func init() {
	cmdMap = map[string]func(*qbot.Client, []string, *qbot.Message){
		"echo":   Echo,
		"rawmsg": Rawmsg,
	}
}

func FindCommand(cmd string) func(*qbot.Client, []string, *qbot.Message) {
	if cmd == "" {
		return nil
	}
	return cmdMap[cmd]
}
