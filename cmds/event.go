package cmds

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go-hurobot/qbot"
)

func cmd_event(c *qbot.Client, raw *qbot.Message, args *ArgsList) {
	const help = `Usage: event [list | del <idx> | clear | msg=<regex> reply=<text> [user=<user_id>] [rand=<0.0-1.0>]]

Examples:
  event list - 查看所有事件
  event del 0 - 删除第 0 个事件
  event clear - 删除所有事件
  event msg=hello reply=world - 添加事件：当消息包含 "hello" 时，回复 "world"
  event msg=".*test.*" reply="matched" rand=0.5 - 添加事件：当消息包含 "test" 时，回复 "matched"，触发概率 50%`

	if args.Size == 1 {
		c.SendMsg(raw, help)
		return
	}

	switch args.Contents[1] {
	case "list":
		listUserEvents(c, raw)
	case "del", "delete":
		if args.Size != 3 {
			c.SendMsg(raw, "Usage: event del <idx>")
			return
		}
		deleteUserEvent(c, raw, args.Contents[2])
	case "clear":
		clearUserEvents(c, raw)
	default:
		// Parse parameters for adding new event
		addUserEvent(c, raw, args)
	}
}

func listUserEvents(c *qbot.Client, msg *qbot.Message) {
	var events []qbot.UserEvents
	result := qbot.PsqlDB.Where("user_id = ?", msg.UserID).
		Order("event_idx").Find(&events)

	if result.Error != nil {
		c.SendMsg(msg, "Database error: "+result.Error.Error())
		return
	}

	if len(events) == 0 {
		c.SendMsg(msg, "No events found")
		return
	}

	var output strings.Builder
	output.WriteString("Your events:\n")
	for _, event := range events {
		output.WriteString(fmt.Sprintf("[%d] msg=%q reply=%q rand=%.2f\n",
			event.EventIdx, event.MsgRegex, event.ReplyText, event.RandProb))
	}
	c.SendMsg(msg, output.String())
}

func deleteUserEvent(c *qbot.Client, msg *qbot.Message, idxStr string) {
	idx, err := strconv.Atoi(idxStr)
	if err != nil || idx < 0 || idx > 9 {
		c.SendMsg(msg, "Invalid index. Must be 0-9")
		return
	}

	result := qbot.PsqlDB.Where("user_id = ? AND event_idx = ?", msg.UserID, idx).
		Delete(&qbot.UserEvents{})

	if result.Error != nil {
		c.SendMsg(msg, "Database error: "+result.Error.Error())
		return
	}

	if result.RowsAffected == 0 {
		c.SendMsg(msg, "Event not found")
		return
	}

	c.SendMsg(msg, fmt.Sprintf("Deleted event %d", idx))
}

func clearUserEvents(c *qbot.Client, msg *qbot.Message) {
	result := qbot.PsqlDB.Where("user_id = ?", msg.UserID).Delete(&qbot.UserEvents{})

	if result.Error != nil {
		c.SendMsg(msg, "Database error: "+result.Error.Error())
		return
	}

	c.SendMsg(msg, fmt.Sprintf("Cleared %d events", result.RowsAffected))
}

func addUserEvent(c *qbot.Client, msg *qbot.Message, args *ArgsList) {
	// Parse key=value parameters
	params := parseEventParams(args)

	msgRegex, ok := params["msg"]
	if !ok {
		c.SendMsg(msg, "Missing required parameter: msg")
		return
	}

	replyText, ok := params["reply"]
	if !ok {
		c.SendMsg(msg, "Missing required parameter: reply")
		return
	}

	// Validate regex
	if _, err := regexp.Compile(msgRegex); err != nil {
		c.SendMsg(msg, "Invalid regex: "+err.Error())
		return
	}

	// Parse optional parameters
	targetUserID := msg.UserID
	if userIDStr, ok := params["user"]; ok {
		if uid := str2uin64(userIDStr); uid != 0 {
			targetUserID = uid
		} else {
			c.SendMsg(msg, "Invalid user ID")
			return
		}
	}

	randProb := float32(1.0)
	if randStr, ok := params["rand"]; ok {
		if prob, err := strconv.ParseFloat(randStr, 32); err != nil || prob < 0 || prob > 1 {
			c.SendMsg(msg, "Invalid rand value. Must be 0.0-1.0")
			return
		} else {
			randProb = float32(prob)
		}
	}

	// Count existing events for this user
	var count int64
	qbot.PsqlDB.Model(&qbot.UserEvents{}).Where("user_id = ?", targetUserID).Count(&count)

	if count >= 10 {
		c.SendMsg(msg, "Maximum 10 events per user")
		return
	}

	// Find next available index
	var existingIndexes []int
	qbot.PsqlDB.Model(&qbot.UserEvents{}).Where("user_id = ?", targetUserID).
		Pluck("event_idx", &existingIndexes)

	nextIdx := 0
	for nextIdx < 10 {
		found := false
		for _, idx := range existingIndexes {
			if idx == nextIdx {
				found = true
				break
			}
		}
		if !found {
			break
		}
		nextIdx++
	}

	// Create new event
	event := qbot.UserEvents{
		UserID:    targetUserID,
		EventIdx:  nextIdx,
		MsgRegex:  msgRegex,
		ReplyText: decodeSpecialChars(replyText),
		RandProb:  randProb,
		CreatedAt: time.Now(),
	}

	if err := qbot.PsqlDB.Create(&event).Error; err != nil {
		c.SendMsg(msg, "Database error: "+err.Error())
		return
	}

	c.SendMsg(msg, fmt.Sprintf("Added event %d: msg=%q reply=%q rand=%.2f",
		nextIdx, msgRegex, replyText, randProb))
}

func parseEventParams(args *ArgsList) map[string]string {
	params := make(map[string]string)

	// Join all arguments starting from index 1
	fullArgs := strings.Join(args.Contents[1:], " ")

	// Parse key=value pairs, handling quoted values
	i := 0
	for i < len(fullArgs) {
		// Skip whitespace
		for i < len(fullArgs) && fullArgs[i] == ' ' {
			i++
		}
		if i >= len(fullArgs) {
			break
		}

		// Find key
		keyStart := i
		for i < len(fullArgs) && fullArgs[i] != '=' && fullArgs[i] != ' ' {
			i++
		}
		if i >= len(fullArgs) || fullArgs[i] != '=' {
			// Not a key=value pair, skip to next space
			for i < len(fullArgs) && fullArgs[i] != ' ' {
				i++
			}
			continue
		}

		key := fullArgs[keyStart:i]
		i++ // skip '='

		// Find value
		valueStart := i
		var value string

		if i < len(fullArgs) && (fullArgs[i] == '"' || fullArgs[i] == '\'') {
			// Quoted value
			quote := fullArgs[i]
			i++ // skip opening quote
			valueStart = i
			for i < len(fullArgs) && fullArgs[i] != quote {
				i++
			}
			value = fullArgs[valueStart:i]
			if i < len(fullArgs) {
				i++ // skip closing quote
			}
		} else {
			// Unquoted value
			for i < len(fullArgs) && fullArgs[i] != ' ' {
				i++
			}
			value = fullArgs[valueStart:i]
		}

		params[key] = value
	}

	return params
}

// CheckUserEvents checks if any user events should trigger for the given message
func CheckUserEvents(c *qbot.Client, msg *qbot.Message) bool {
	var events []qbot.UserEvents
	result := qbot.PsqlDB.Where("user_id = ?", msg.UserID).Find(&events)

	if result.Error != nil {
		return false
	}

	triggered := false
	for _, event := range events {
		// Check if regex matches
		if regex, err := regexp.Compile(event.MsgRegex); err == nil {
			if regex.MatchString(msg.Content) || regex.MatchString(msg.Raw) {
				// Check probability
				if event.RandProb >= 1.0 || rand.Float32() <= event.RandProb {
					c.SendMsg(msg, event.ReplyText)
					triggered = true
				}
			}
		}
	}

	return triggered
}
