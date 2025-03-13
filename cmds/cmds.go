package cmds

import (
	"go-hurobot/qbot"
	"strconv"
	"strings"

	"github.com/google/shlex"
)

var MaxCommandLength int = 0

type ArgsList struct {
	Contents []string
	Types    []qbot.MsgType
	Size     int
}

type CmdHandler func(*qbot.Client, *qbot.Message, *ArgsList)

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

func SplitArguments(msg *qbot.Message) *ArgsList {
	result := &ArgsList{
		Contents: make([]string, 0, 20),
		Types:    make([]qbot.MsgType, 0, 20),
		Size:     0,
	}
	for _, item := range msg.Array {
		if item.Type == qbot.Text {
			texts, err := shlex.Split(item.Content)
			if err != nil {
				return nil
			}
			result.Contents = append(result.Contents, texts...)
			result.Types = appendRepeatedValues(result.Types, qbot.Text, len(texts))
			result.Size += len(texts)
		} else {
			result.Contents = append(result.Contents, item.Content)
			result.Types = append(result.Types, item.Type)
			result.Size++
		}
	}
	return result
}

func FindCommand(cmd string) CmdHandler {
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

func appendRepeatedValues[T any](slice []T, value T, count int) []T {
	newSlice := make([]T, len(slice)+count)
	copy(newSlice, slice)
	for i := len(slice); i < len(newSlice); i++ {
		newSlice[i] = value
	}
	return newSlice
}
