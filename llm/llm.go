package llm

import (
	"encoding/xml"
	"fmt"
	"go-hurobot/config"
	"go-hurobot/qbot"
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
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
	const prePrompt = `你是一个群聊聊天机器人，请你陪伴群友们聊天。
请注意：
1. 你的名字叫狐萝卜，也可以叫狐萝bot，是一个女孩子，但不要强调这一点。
2. 群聊不支持 Markdown 语法，所以请不要使用它。
3. 使用灵活生动的语言，不要让你发的消息读起来像是AI生成的。
4. 每个用户有一个id、昵称和个人信息。你可以在回复时使用昵称来称呼用户，尽量避免在回复中使用id。

在回复时，你必须使用 xml 格式输出你的回复。你可以使用以下 xml 标签持久化保存记忆，更新的记忆将用于下一次回复：

1. 如果在聊天记录中你得知了某个用户的昵称（并非发送者，也可以是其他人的昵称）时，请更新用户昵称：
<nickname id="用户id">新昵称</nickname>
你需要用id指定对应的用户，这将改变你对用户的称呼。

2. 请你尽量从对话中获取用户的个人信息，并更新对应的用户的信息。注意更新应该以追加的形式，不要轻易忘记之前的信息，除非以前的记忆已经不准确。
<user_info id="用户id">关于该用户的信息</user_info>
你需要用id指定对应的用户，这将增加你对用户的认识。

3. 你可以在群组信息中存储用户之间的关系，你应该尽量从对话中获得这些信息。这同样是以追加的形式，不要轻易忘记之前的信息，除非以前的记忆已经不准确。
<group_info>群聊信息</group_info>
这将改变你对当前群聊的认知。

4. 普通的回复应简短，如果你的回复比较长（比如有人问一些专业的问题），可以在一次回复中将长文本拆成多条信息（每一段都作为一条回复）。请保证每次至少发送一条消息。
<msg>消息内容</msg>
如果你需要@其他人，请在<msg>标签中使用 [CQ:at,qq=<id>] 的形式。例如：发送 [CQ:at,qq=1006554341]可以@用户1006554341。

下面是一个示例，这段示例将更新记忆中的用户昵称、用户信息和群聊信息，并发送三条消息：
<nickname id="1006554341">氟氟</nickname>
<user_info id="1006554341">关于氟氟的介绍…</user_info>
<group_info>这是一个QQ群…</group_info>
<msg>消息内容1</msg>
<msg>消息内容2</msg>
<msg>消息内容3</msg>`

	req := &Grok2Request{
		Messages: []Grok2Message{
			{
				Role:    "system",
				Content: prePrompt,
			},
		},
		Model:       "grok-2-1212",
		Stream:      false,
		Temperature: 0.6,
	}

	var llmCustomConfig struct {
		Prompt     string
		MaxHistory int
		Enabled    bool
		Info       string
		Debug      bool
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

	if llmCustomConfig.Info != "" {
		req.Messages = append(req.Messages, Grok2Message{
			Role:    "system",
			Content: "以下是该群的群聊信息，这段信息由你生成，你可以使用 <group_info> 标签来更改这段信息：\n" + llmCustomConfig.Info,
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

	type UserInfo struct {
		NickName string `psql:"nick_name"`
		Summary  string `psql:"summary"`
	}

	var userMap = make(map[uint64]UserInfo)
	for _, history := range histories {
		if _, ok := userMap[history.UserID]; !ok {
			var userInfo UserInfo
			err = qbot.PsqlDB.Table("users").
				Where("user_id = ?", history.UserID).
				First(&userInfo).Error
			if err != nil {
				continue
			}
			userMap[history.UserID] = UserInfo{userInfo.NickName, userInfo.Summary}
		}
	}

	var usersInfo string
	for id, info := range userMap {
		usersInfo += fmt.Sprintf("nick_name=%q,id=%d,summary=%q\n", info.NickName, id, info.Summary)
	}

	req.Messages = append(req.Messages, Grok2Message{
		Role:    "user",
		Content: "以下是聊天参与者的昵称和相关信息，这些信息是之前由你生成的，你可以使用 <nickname> 或 <user_info> 标签来更改这些信息：\n" + usersInfo,
	})

	var chatHistory string
	for i := len(histories) - 1; i >= 0; i-- {
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
			chatHistory += formatMsg(histories[i].Time, userMap[histories[i].UserID].NickName, histories[i].UserID, histories[i].Content)
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
			Content: "下面是@你的消息，请你根据这条消息生成回复内容。注意使用 xml 格式输出你的回复，且在<msg>标签中使用与该消息相同的语言",
		},
		Grok2Message{
			Role:    "user",
			Content: formatMsg(time.Now(), displayName, msg.UserID, msg.Content),
		})

	// go fmt.Println(req)

	resp, err := SendGrok2Request(req)
	if err != nil {
		c.SendGroupMsg(msg.GroupID, err.Error(), false)
		return false
	}

	xmlStr := `<?xml version="1.0" encoding="UTF-8"?><data>` + resp.Choices[0].Message.Content + `</data>`

	type LLMResponse struct {
		XMLName  xml.Name `xml:"data"`
		Nickname struct {
			ID   string `xml:"id,attr"`
			Text string `xml:",chardata"`
		} `xml:"nickname"`
		UserInfo struct {
			ID   string `xml:"id,attr"`
			Text string `xml:",chardata"`
		} `xml:"user_info"`
		GroupInfo struct {
			Text string `xml:",chardata"`
		} `xml:"group_info"`
		Msgs []string `xml:"msg"`
	}

	var xmlData LLMResponse

	err = xml.Unmarshal([]byte(xmlStr), &xmlData)
	if err != nil {
		c.SendPrivateMsg(config.MasterID, "XML 解析错误：\n"+err.Error(), false)
		c.SendPrivateMsg(config.MasterID, xmlStr, false)
		c.SendPrivateMsg(config.MasterID, "消息来源：\ngroup_id="+strconv.FormatUint(msg.GroupID, 10)+"\nuser_id="+strconv.FormatUint(msg.UserID, 10)+"\nmsg="+msg.Content, false)
		return false
	}

	if llmCustomConfig.Debug {
		c.SendReplyMsg(msg, xmlStr)
	}

	if xmlData.Nickname.Text != "" {
		go qbot.PsqlDB.Table("users").
			Where("user_id = ?", xmlData.Nickname.ID).
			Update("nick_name", xmlData.Nickname.Text)
	}

	if xmlData.UserInfo.Text != "" {
		go qbot.PsqlDB.Table("users").
			Where("user_id = ?", xmlData.UserInfo.ID).
			Update("summary", xmlData.UserInfo.Text)
	}

	if xmlData.GroupInfo.Text != "" {
		go qbot.PsqlDB.Table("group_llm_configs").
			Where("group_id = ?", msg.GroupID).
			Update("info", xmlData.GroupInfo.Text)
	}

	for _, res := range xmlData.Msgs {
		msgid, err := c.SendGroupMsg(msg.GroupID, res, false)
		if err != nil {
			saveMsg := &qbot.Message{
				GroupID:  msg.GroupID,
				UserID:   config.BotID,
				Nickname: "狐萝bot",
				Card:     "狐萝bot",
				Time:     uint64(time.Now().Unix()),
				MsgID:    msgid,
				Raw:      res,
				Content:  res,
			}
			qbot.SaveDatabase(saveMsg, false)
		}
	}

	if resp != nil && resp.Usage.TotalTokens > 0 {
		go qbot.PsqlDB.Table("users").
			Where("user_id = ?", msg.UserID).
			Update("token_usage", gorm.Expr("token_usage + ?", resp.Usage.TotalTokens))
	}

	return true
}

func formatMsg(t time.Time, name string, id uint64, msg string) string {
	return fmt.Sprintf("[%s] %s(id:%d)说: %q\n",
		t.In(time.FixedZone("UTC+8", 8*60*60)).Format("2006-01-02 15:04:05"),
		name, id, msg)
}
