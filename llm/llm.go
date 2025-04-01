package llm

import (
	"fmt"
	"go-hurobot/config"
	"go-hurobot/qbot"
	"log"
	"strconv"
	"time"
)

func LLMMsgHandle(c *qbot.Client, msg *qbot.Message) bool {
	reply := false
	for _, item := range msg.Array {
		if item.Type == qbot.At && item.Content == strconv.FormatUint(config.BotID, 10) {
			reply = true
		}
	}
	if !reply {
		return false
	}
	const prePrompt = `你是一个群聊聊天机器人，请你陪伴群友们聊天：
		你需要根据用户@你的消息进行回复。你能在用户提示词中看到一些群聊历史记录，你可以参考这些聊天记录进行回复。
		如果你需要@其他人，请使用 [CQ:at,qq=<id>] 的形式。例如：发送 [CQ:at,qq=1006554341] 可以@用户1006554341。
		以下是一些注意事项：
		1. 群聊不支持 Markdown 语法，所以请不要使用它。
		2. 使用灵活生动的语言，不要让你发的消息读起来像是AI生成的。
		3. 由于你在一个群聊中，所以你给出的响应要尽量简短，以表现得更像人类。`

	req := &Grok2Request{
		Messages: []Grok2Message{
			{
				Role:    "system",
				Content: prePrompt,
			},
		},
		Model:       "grok-2-1212",
		Stream:      false,
		Temperature: 0.5,
	}

	var llmCustomConfig struct {
		Prompt     string
		MaxHistory int
		Enabled    bool
	}

	err := qbot.PsqlDB.Table("group_llm_configs").
		Where("group_id = ?", msg.GroupID).
		First(&llmCustomConfig).Error

	if err != nil || !llmCustomConfig.Enabled {
		return false
	}

	if llmCustomConfig.Prompt != "" {
		req.Messages = append(req.Messages, Grok2Message{
			Role:    "system",
			Content: llmCustomConfig.Prompt,
		})
	}

	var histories []struct {
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
		Limit(llmCustomConfig.MaxHistory).
		Find(&histories).Error

	if err != nil {
		log.Println(err.Error())
		return false
	}

	var chatHistory string
	for i := len(histories) - 1; i >= 0; i-- {
		displayName := histories[i].Name
		if histories[i].Nickname != "" {
			displayName = histories[i].Nickname
		}
		localTime := histories[i].Time.In(time.FixedZone("UTC+8", 8*60*60))
		if histories[i].UserID == config.BotID {
			if chatHistory != "" {
				req.Messages = append(req.Messages, Grok2Message{
					Role:    "user",
					Content: chatHistory,
				})
				chatHistory = ""
			}
			req.Messages = append(req.Messages, Grok2Message{
				Role:    "assistant",
				Content: localTime.Format("2006-01-02 15:04:05 ") + histories[i].Content,
			})
		} else {
			chatHistory += formatMsg(histories[i].Time, displayName, histories[i].UserID, histories[i].Content)
		}
	}
	if chatHistory != "" {
		req.Messages = append(req.Messages, Grok2Message{
			Role:    "user",
			Content: chatHistory,
		})
	}

	var userInfo struct {
		NickName string
	}
	err = qbot.PsqlDB.Table("users").
		Select("nick_name").
		Where("user_id = ?", msg.UserID).
		First(&userInfo).Error

	displayName := msg.Card
	if err == nil && userInfo.NickName != "" {
		displayName = userInfo.NickName
	}

	req.Messages = append(req.Messages,
		Grok2Message{
			Role:    "system",
			Content: "下面是@你的消息，请你根据这条消息生成回复内容。使用与该消息相同的语言，不需要带时间。",
		},
		Grok2Message{
			Role:    "user",
			Content: formatMsg(time.Now(), displayName, msg.UserID, msg.Content),
		})

	go fmt.Println(req)

	resp, err := SendGrok2Request(req)
	if err != nil {
		c.SendGroupMsg(msg.GroupID, err.Error(), false)
		return false
	}

	if len(resp.Choices) > 0 {
		msgid, err := c.SendGroupMsg(msg.GroupID, resp.Choices[0].Message.Content, false)
		if err != nil {
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
	}
	return true
}

func formatMsg(t time.Time, name string, id uint64, msg string) string {
	return fmt.Sprintf("[%s] %s(id:%d)说: %q\n",
		t.In(time.FixedZone("UTC+8", 8*60*60)).Format("2006-01-02 15:04:05"),
		name, id, msg)
}
