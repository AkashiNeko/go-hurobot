package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"go-hurobot/config"
	"go-hurobot/llm"
	"go-hurobot/qbot"
)

func customReply(c *qbot.Client, msg *qbot.Message) {
	if msg.Array[0].Type == qbot.At && msg.Array[0].Content == strconv.FormatUint(config.BotID, 10) {
		var llmConfig struct {
			Prompt     string
			MaxHistory int
			Enabled    bool
		}
		err := qbot.PsqlDB.Table("group_llm_configs").
			Where("group_id = ?", msg.GroupID).
			First(&llmConfig).Error

		if err != nil || !llmConfig.Enabled {
			return
		}

		var messages []struct {
			UserID   uint64
			Content  string
			Name     string
			Nickname string
			Time     time.Time
		}

		err = qbot.PsqlDB.Table("messages").
			Select("messages.user_id, messages.content, users.name, users.nick_name, messages.time").
			Joins("LEFT JOIN users ON messages.user_id = users.user_id").
			Where("messages.group_id = ? AND messages.is_cmd = false", msg.GroupID).
			Order("messages.time DESC").
			Limit(llmConfig.MaxHistory).
			Find(&messages).Error

		if err != nil {
			log.Println(err.Error())
			return
		}

		var chatHistory string
		for i := len(messages) - 1; i >= 0; i-- {
			displayName := messages[i].Name
			if messages[i].Nickname != "" {
				displayName = messages[i].Nickname
			}
			if messages[i].UserID == config.BotID {
				chatHistory += fmt.Sprintf("你自己说: %q\n", messages[i].Content)
			} else {
				chatHistory += fmt.Sprintf("%s(%d)说: %q\n",
					displayName,
					messages[i].UserID,
					messages[i].Content)
			}
		}

		grok2Req := &llm.Grok2Request{
			Messages: []llm.Grok2Message{
				{
					Role:    "system",
					Content: llmConfig.Prompt,
				},
				{
					Role:    "user",
					Content: chatHistory,
				},
				{
					Role:    "system",
					Content: "下面是@你的消息，请你根据这条消息生成回复内容。",
				},
				{
					Role:    "user",
					Content: fmt.Sprintf("%s(%d)说: %q", msg.Card, msg.UserID, msg.Content),
				},
			},
			Model:       "grok-2-1212",
			Stream:      false,
			Temperature: 0.5,
		}

		resp, err := llm.SendGrok2Request(grok2Req)
		if err != nil {
			c.SendGroupMsg(msg.GroupID, err.Error(), false)
			return
		}

		if len(resp.Choices) > 0 {
			msgid, _ := c.SendGroupMsg(msg.GroupID, resp.Choices[0].Message.Content, false)
			saveMsg := &qbot.Message{
				GroupID:  msg.GroupID,
				UserID:   config.BotID,
				Nickname: "狐萝bot",
				Card:     "狐萝bot",
				Role:     "member",
				Time:     uint64(time.Now().Unix()),
				MsgID:    msgid,
				Raw:      resp.Choices[0].Message.Content,
				Content:  resp.Choices[0].Message.Content,
				Array:    nil,
			}
			qbot.SaveDatabase(saveMsg, false)
		}
		return
	}

	// 2025-03-08 晚上，让 bot 在某 mc 群发电加的
	if msg.GroupID == 158045531 {
		if strings.Contains(msg.Raw, "厉厉厉害害害") {
			return
		} else if strings.Contains(msg.Raw, "厉厉害害") {
			c.SendGroupMsg(msg.GroupID, strings.Replace(msg.Raw, "厉厉害害", "可可爱爱", -1), false)
		} else if strings.Contains(msg.Raw, "厉害害") {
			c.SendGroupMsg(msg.GroupID, strings.Replace(msg.Raw, "厉害害", "可爱爱", -1), false)
		} else if strings.Contains(msg.Raw, "厉害") {
			c.SendGroupMsg(msg.GroupID, strings.Replace(msg.Raw, "厉害", "可爱", -1), false)
		}
	}
	// 2025-03-08 ↑
}
