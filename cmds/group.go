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
	const help = "Usage: group [rename <group name> | op [@user1 @user2 ...] | deop [@user1 @user2 ...] | banme <time> | ban @user <time>]"
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
		setGroupAdmin(c, raw, args, true)
	case "deop":
		setGroupAdmin(c, raw, args, false)
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
	case "ban":
		if args.Size != 4 {
			c.SendMsg(raw, help)
		} else if args.Types[2] == qbot.At {
			time, err := strconv.Atoi(args.Contents[3])
			if err != nil || time < 1 || time > 24*60*30 {
				c.SendMsg(raw, "Invalid time duration")
				return
			}
			c.SetGroupBan(raw.GroupID, str2uin64(args.Contents[2]), time*60)
		} else {
			c.SendMsg(raw, "Invalid user")
		}
	}
}

func setGroupAdmin(c *qbot.Client, raw *qbot.Message, args *ArgsList, isOp bool) {
	targetUserIDs, err := extractTargetUsers(args, 2, raw.UserID)
	if err != nil {
		c.SendMsg(raw, "Invalid argument: "+err.Error())
		return
	}

	validUserIDs := make([]uint64, 0, len(targetUserIDs))
	userIDSet := make(map[uint64]bool)

	action := "op"
	if !isOp {
		action = "deop"
	}

	for _, userID := range targetUserIDs {
		if userID == config.BotID {
			c.SendMsg(raw, fmt.Sprintf("Cannot %s bot", action))
			continue
		}
		if !userIDSet[userID] {
			userIDSet[userID] = true
			validUserIDs = append(validUserIDs, userID)
			if len(validUserIDs) >= 10 {
				break
			}
		}
	}

	if len(validUserIDs) == 0 {
		return
	}

	for _, userID := range validUserIDs {
		c.SetGroupAdmin(raw.GroupID, userID, isOp)
	}

	if len(validUserIDs) == 1 {
		c.SendMsg(raw, fmt.Sprintf("%s: %d", action, validUserIDs[0]))
	} else {
		userIDStrings := make([]string, len(validUserIDs))
		for i, id := range validUserIDs {
			userIDStrings[i] = strconv.FormatUint(id, 10)
		}
		c.SendMsg(raw, fmt.Sprintf("%s: %s", action, strings.Join(userIDStrings, ", ")))
	}
}

func extractTargetUsers(args *ArgsList, startIndex int, defaultUserID uint64) ([]uint64, error) {
	var targetUserIDs []uint64
	hasAtUsers := false

	for i := startIndex; i < args.Size; i++ {
		if args.Types[i] == qbot.At {
			hasAtUsers = true
			targetUserIDs = append(targetUserIDs, str2uin64(args.Contents[i]))
		} else {
			return nil, fmt.Errorf("use @ to mention users")
		}
	}

	if !hasAtUsers {
		targetUserIDs = append(targetUserIDs, defaultUserID)
	}

	return targetUserIDs, nil
}
