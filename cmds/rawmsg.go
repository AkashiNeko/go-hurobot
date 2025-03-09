package cmds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-hurobot/qbot"
)

func cmd_rawmsg(c *qbot.Client, args []string, raw *qbot.Message) {
	if len(args) >= 2 && (args[1] == "-f" || args[1] == "--format") {
		if len(args) >= 3 {
			switch args[2] {
			case "json": // default
			case "%v":
				fallthrough
			case "%+v":
				fallthrough
			case "%#v":
				c.SendReplyMsg(raw, fmt.Sprintf(args[2], raw))
				return
			default:
				c.SendReplyMsg(raw, fmt.Sprintf("Unknown format %q", args[2]))
				return
			}
		} else {
			c.SendReplyMsg(raw, fmt.Sprintf("Usage: %s [-f|--format format]", args[0]))
			return
		}
	}
	jsonStr, _ := json.Marshal(raw)
	jsonBytes := []byte(jsonStr)
	var out bytes.Buffer
	err := json.Indent(&out, jsonBytes, "", "  ")
	if err != nil {
		c.SendReplyMsg(raw, string(jsonStr))
	} else {
		c.SendReplyMsg(raw, out.String())
	}
}
