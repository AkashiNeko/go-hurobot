package cmds

import (
	"fmt"
	"go-hurobot/qbot"
	"strings"
)

func cmd_group(c *qbot.Client, raw *qbot.Message, args *ArgsList) {
	if raw.GroupID != 948697448 {
		return
	}
	const help = "Usage: group [rename <group name>]"
	if args.Size == 1 {
		c.SendMsg(raw, help)
	}
	switch args.Contents[1] {
	case "rename":
		if args.Size < 3 {
			c.SendMsg(raw, help)
		} else {
			newName := decodeSpecialChars(strings.Join(args.Contents[2:], " "))
			c.SendMsg(raw, fmt.Sprintf("重命名群名: %q", newName))
			c.SetGroupName(raw.GroupID, newName)
		}
	case "op":
		c.SendMsg(raw, fmt.Sprintf("已将 %s 设为 WTF 管理员。", raw.Card))
		c.SetGroupAdmin(raw.GroupID, raw.UserID, true)
	case "deop":
		c.SendMsg(raw, fmt.Sprintf("已取消 %s 的 WTF 管理员身份。", raw.Card))
		c.SetGroupAdmin(raw.GroupID, raw.UserID, false)
	}
}
