package cmds

import (
	"go-hurobot/qbot"
	"strconv"
	"strings"
)

var MaxCommandLength int = 0

type CmdHandler func(*qbot.Client, []string, *qbot.Message)

var cmdMap map[string]CmdHandler

func init() {
	cmdMap = map[string]CmdHandler{
		"echo":         cmd_echo,
		"rawmsg":       cmd_rawmsg,
		"grok2":        cmd_grok2,
		"specialtitle": cmd_specialtitle,
		"sh":           cmd_sh,
		"psql":         cmd_psql,
		"group":        cmd_group,
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

func decodeSpecialChars(raw string) string {
	replacer := strings.NewReplacer(
		"&#91;", "[",
		"&#93;", "]",
		"&amp;", "&",
	)
	return replacer.Replace(raw)
}

func str2uin64(s string) uint64 {
	value, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return value
}
