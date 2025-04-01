package cmds

import (
	"go-hurobot/qbot"
	"strconv"
)

func cmd_callme(c *qbot.Client, msg *qbot.Message, args *ArgsList) {
	var targetID uint64
	var nickname string
	var isQuery bool = true

	switch args.Size {
	case 1: // callme
		targetID = msg.UserID
	case 2: // callme <nickname> / callme <@id>
		if args.Types[1] == qbot.At {
			// callme <@id>
			targetID = str2uin64(args.Contents[1])
		} else {
			// callme <nickname>
			targetID = msg.UserID
			nickname = args.Contents[1]
			isQuery = false
		}
	case 3: // callme <@id> <nickname> 或 callme <nickname> <@id>
		isQuery = false
		if args.Types[1] == qbot.At {
			// callme <@id> <nickname>
			targetID = str2uin64(args.Contents[1])
			nickname = args.Contents[2]
		} else if args.Types[2] == qbot.At {
			// callme <nickname> <@id>
			targetID = str2uin64(args.Contents[2])
			nickname = args.Contents[1]
		} else {
			return
		}
	default:
		c.SendMsg(msg, `Usage:
		- callme
		- callme <nickname>
		- callme <@id>
		- callme <@id> <nickname>`)
		return
	}

	if isQuery {
		var user qbot.Users
		result := qbot.PsqlDB.Where("user_id = ?", targetID).First(&user)
		if result.Error != nil || user.Nickname == "" {
			c.SendMsg(msg, "")
			return
		}
		c.SendMsg(msg, user.Nickname)
	} else {
		user := qbot.Users{
			UserID:   targetID,
			Name:     msg.Nickname,
			Nickname: nickname,
		}

		result := qbot.PsqlDB.Where("user_id = ?", targetID).Assign(
			qbot.Users{Nickname: nickname},
		).FirstOrCreate(&user)

		if result.Error != nil {
			c.SendMsg(msg, "failed")
			return
		}
		if targetID == msg.UserID {
			c.SendMsg(msg, "Update nickname: "+nickname)
		} else {
			c.SendMsg(msg, "Update nickname for "+strconv.FormatUint(targetID, 10)+": "+nickname)
		}
	}
}
