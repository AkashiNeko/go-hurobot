package cmds

import (
	"fmt"
	"go-hurobot/qbot"
	"time"
)

func cmd_memberinfo(c *qbot.Client, msg *qbot.Message, args *ArgsList) {
	// 只能在群聊中使用
	if msg.GroupID == 0 {
		return
	}

	var targetUserID uint64

	if args.Size >= 2 && args.Types[1] == qbot.At {
		targetUserID = str2uin64(args.Contents[1])
	} else {
		targetUserID = msg.UserID
	}

	if targetUserID == 0 {
		c.SendReplyMsg(msg, "Invalid user ID")
		return
	}

	// 获取群成员信息
	memberInfo, err := c.GetGroupMemberInfo(msg.GroupID, targetUserID, false)
	if err != nil {
		c.SendReplyMsg(msg, fmt.Sprintf("Failed to get member info: %v", err))
		return
	}

	response := fmt.Sprintf(
		"QQ号: %d\n"+
			"昵称: %s\n"+
			"名片: %s\n"+
			"性别: %s\n"+
			"权限: %s\n"+
			"等级: Lv %s",
		memberInfo.UserID,
		memberInfo.Nickname,
		getCardOrNickname(memberInfo.Card, memberInfo.Nickname),
		getSexString(memberInfo.Sex),
		getRoleString(memberInfo.Role),
		memberInfo.Level)

	if memberInfo.Age > 0 {
		response += fmt.Sprintf("\n年龄: %d", memberInfo.Age)
	}

	if memberInfo.Area != "" {
		response += fmt.Sprintf("\n地区: %s", memberInfo.Area)
	}

	if memberInfo.Title != "" {
		response += fmt.Sprintf("\n头衔: %s", memberInfo.Title)
	}

	if memberInfo.ShutUpTimestamp > 0 {
		shutUpTime := time.Unix(memberInfo.ShutUpTimestamp, 0)
		if shutUpTime.After(time.Now()) {
			response += fmt.Sprintf("\n禁言到期: %s", shutUpTime.Format("2006-01-02 15:04:05"))
		}
	}

	if memberInfo.JoinTime > 0 {
		joinTime := time.Unix(int64(memberInfo.JoinTime), 0)
		response += fmt.Sprintf("\n加群时间: %s", joinTime.Format("2006-01-02 15:04:05"))
	}

	c.SendReplyMsg(msg, response)
}

func getCardOrNickname(card, nickname string) string {
	if card != "" {
		return card
	}
	return nickname
}

func getSexString(sex string) string {
	switch sex {
	case "male":
		return "♂"
	case "female":
		return "♀"
	default:
		return "?"
	}
}

func getRoleString(role string) string {
	switch role {
	case "owner":
		return "👑群主"
	case "admin":
		return "管理员"
	case "member":
		return "🐱成员"
	default:
		return role
	}
}
