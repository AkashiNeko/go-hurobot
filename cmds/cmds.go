package cmds

import (
	"go-hurobot/qbot"
	"strconv"
	"strings"

	"github.com/google/shlex"
)

var maxCommandLength int = 0

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
		"delete":       cmd_delete,
	}

	for key := range cmdMap {
		if len(key) > maxCommandLength {
			maxCommandLength = len(key)
		}
	}
}

func HandleCommand(c *qbot.Client, msg *qbot.Message) {
	skip := 0
	if msg.Array[0].Type == qbot.Reply {
		skip++
		if len(msg.Array) != 1 && msg.Array[1].Type == qbot.At {
			skip++
		}
	} else if msg.Array[0].Type == qbot.At {
		skip++
	}

	if skip != 0 {
		if p := findNthClosingBracket(msg.Raw, skip); p != len(msg.Raw) {
			msg.Raw = msg.Raw[p:]
		} else {
			return
		}
	}
	handler := findCommand(getCommandName(msg.Raw))
	go qbot.SaveDatabase(msg, handler != nil)
	if handler != nil {
		if args := splitArguments(msg, skip); args != nil {
			handler(c, msg, args)
		}
	}
}

func splitArguments(msg *qbot.Message, skip int) *ArgsList {
	result := &ArgsList{
		Contents: make([]string, 0, 20),
		Types:    make([]qbot.MsgType, 0, 20),
		Size:     0,
	}

	if skip < 0 {
		skip = 0
	}

	if skip >= len(msg.Array) {
		return result
	}

	for _, item := range msg.Array[skip:] {
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

func findNthClosingBracket(s string, n int) int {
	count := 0
	for i, char := range s {
		if char == ']' {
			count++
			if count == n {
				i++
				for i < len(s) && s[i] == ' ' {
					i++
				}
				return i
			}
		}
	}
	return 0
}

func findCommand(cmd string) CmdHandler {
	if cmd == "" {
		return nil
	}
	return cmdMap[cmd]
}

func getCommandName(s string) string {
	sliced := false
	if len(s) > maxCommandLength+1 {
		s = s[:maxCommandLength+1]
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

func decodeSpecialChars(raw string) string {
	replacer := strings.NewReplacer(
		"&#91;", "[",
		"&#93;", "]",
		"&amp;", "&",
	)
	return replacer.Replace(raw)
}

func encodeSpecialChars(raw string) string {
	replacer := strings.NewReplacer(
		"[", "&#91;",
		"]", "&#93;",
		"&", "&amp;",
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
