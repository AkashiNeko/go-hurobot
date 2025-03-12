package cmds

import (
	"fmt"
	"go-hurobot/qbot"
	"strings"
)

func cmd_group(c *qbot.Client, args []string, raw *qbot.Message) {
	if raw.GroupID == 0 {
		c.SendReplyMsg(raw, "只能在群组中使用")
		return
	}
	const help = "Usage: group [rename <group name>]"
	if len(args) == 1 {
		c.SendReplyMsg(raw, help)
	}
	switch args[1] {
	case "rename":
		if len(args) < 3 {
			c.SendReplyMsg(raw, help)
		} else {
			newName := decodeSpecialChars(strings.Join(args[2:], " "))
			c.SendReplyMsg(raw, fmt.Sprintf("重命名群名: %q", newName))
			c.SetGroupName(raw.GroupID, newName)
		}
	case "op":
		// if len(args) < 3 {
		// }
		c.SendReplyMsg(raw, fmt.Sprintf("已将 %s 设为 WTF 管理员。", raw.Sender.Card))
		c.SetGroupAdmin(raw.GroupID, raw.Sender.UserID, true)
	case "deop":
		c.SendReplyMsg(raw, fmt.Sprintf("已取消 %s 的 WTF 管理员身份。", raw.Sender.Card))
		c.SetGroupAdmin(raw.GroupID, raw.Sender.UserID, false)
	}
}
