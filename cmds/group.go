package cmds

import (
	"fmt"
	"go-hurobot/config"
	"go-hurobot/qbot"
	"slices"
	"strconv"
	"strings"
)

func cmd_group(c *qbot.Client, raw *qbot.Message, args *ArgsList) {
	if !slices.Contains(config.BotOwnerGroupIDs, raw.GroupID) {
		return
	}
	const help = "Usage: group [rename <group name> | op | deop | banme <time>]"
	if args.Size == 1 {
		c.SendMsg(raw, help)
		return
	}
	switch args.Contents[1] {
	case "rename":
		if args.Size < 3 {
			c.SendMsg(raw, help)
		} else {
			newName := decodeSpecialChars(strings.Join(args.Contents[2:], " "))
			c.SendMsg(raw, fmt.Sprintf("rename: %q", newName))
			c.SetGroupName(raw.GroupID, newName)
		}
	case "op":
		c.SetGroupAdmin(raw.GroupID, raw.UserID, true)
	case "deop":
		c.SetGroupAdmin(raw.GroupID, raw.UserID, false)
	case "banme":
		if args.Size != 3 {
			c.SendMsg(raw, help)
		} else {
			time, err := strconv.Atoi(args.Contents[2])
			if err != nil || time < 1 || time > 24*60*30 {
				c.SendMsg(raw, "Invalid time duration")
				return
			}
			c.SetGroupBan(raw.GroupID, raw.UserID, time*60)
		}
	}
}
