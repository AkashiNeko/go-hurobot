package cmds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-hurobot/qbot"
)

func cmd_rawmsg(c *qbot.Client, raw *qbot.Message, args *ArgsList) {
	if args.Size >= 2 && (args.Contents[1] == "-f" || args.Contents[1] == "--format") {
		if args.Size >= 3 {
			switch args.Contents[2] {
			case "json": // default
			case "%v":
				fallthrough
			case "%+v":
				fallthrough
			case "%#v":
				c.SendReplyMsg(raw, fmt.Sprintf(args.Contents[2], raw))
				return
			default:
				c.SendReplyMsg(raw, fmt.Sprintf("Unknown format %q", args.Contents[2]))
				return
			}
		} else {
			c.SendReplyMsg(raw, fmt.Sprintf("Usage: %s [-f|--format format]", args.Contents[0]))
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
