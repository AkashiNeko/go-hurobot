package qbot

import (
	"encoding/json"
	"log"
)

func parseContent(msgarr *[]MsgItem) (result string) {
	result = ""
	for _, item := range *msgarr {
		switch item.Type {
		case Text:
			result += item.Content
		case At:
			result += "[CQ:at,qq=" + item.Content + "]"
		case Face:
			faceName := getQFaceNameByStrID(item.Content)
			log.Println(faceName)
			result += "[CQ:face,id=" + item.Content + "](" + faceName + ")"
		case Image:
			result += "[图片(无法查看)]"
		case Record:
			result = "[语音(无法查看)]"
		case Reply:
			result = "[CQ:reply,id=" + item.Content + "]"
		case File:
			result = "[文件(无法查看)]"
		case Forward:
			result = "[合并转发(无法查看)]"
		case Json:
			result = "[不支持的消息(无法查看)]"
		}
	}
	return
}

func parseMsgJson(raw *messageJson) *Message {
	if raw == nil {
		return nil
	}
	result := Message{
		GroupID:  raw.GroupID,
		UserID:   raw.Sender.UserID,
		Nickname: raw.Sender.Nickname,
		Card:     raw.Sender.Card,
		Role:     raw.Sender.Role,
		Time:     raw.Time,
		MsgID:    raw.MessageID,
		Raw:      raw.RawMessage,
	}
	for _, msg := range raw.Message {
		var jsonData map[string]any
		if json.Unmarshal([]byte(msg.Data), &jsonData) != nil {
			return nil
		}
		switch msg.Type {
		case "text":
			result.Array = append(result.Array, MsgItem{
				Type:    Text,
				Content: jsonData["text"].(string),
			})
		case "at":
			result.Array = append(result.Array, MsgItem{
				Type:    At,
				Content: jsonData["qq"].(string),
			})
		case "face":
			result.Array = append(result.Array, MsgItem{
				Type:    Face,
				Content: jsonData["id"].(string),
			})
		case "image":
			result.Array = append(result.Array, MsgItem{
				Type:    Image,
				Content: jsonData["url"].(string),
			})
		case "record":
			result.Array = append(result.Array, MsgItem{
				Type:    Record,
				Content: jsonData["path"].(string),
			})
		case "reply":
			result.Array = append(result.Array, MsgItem{
				Type:    Reply,
				Content: jsonData["id"].(string),
			})
		case "file":
			result.Array = append(result.Array, MsgItem{
				Type:    File,
				Content: string(msg.Data),
			})
		case "forward":
			result.Array = append(result.Array, MsgItem{
				Type:    Forward,
				Content: string(msg.Data),
			})
		case "json":
			result.Array = append(result.Array, MsgItem{
				Type:    Json,
				Content: string(msg.Data),
			})
		default:
			result.Array = append(result.Array, MsgItem{
				Type:    Other,
				Content: string(msg.Data),
			})
		}
	}
	result.Content = parseContent(&result.Array)
	log.Println(result.Content)
	return &result
}

func (c *Client) handleEvents(postType *string, msgStr *[]byte, jsonMap *map[string]any) {
	switch *postType {
	case "meta_event":
		// heartbeat, connection state..
	case "notice":
		// TODO
		switch (*jsonMap)["notice_type"] {
		case "group_recall":
			// TODO
		}
	case "message":
		switch (*jsonMap)["message_type"] {
		case "private":
			fallthrough
		case "group":
			if c.eventHandlers.onMessage != nil {
				msgJson := &messageJson{}
				if json.Unmarshal(*msgStr, msgJson) != nil {
					return
				}
				if msg := parseMsgJson(msgJson); msg != nil {
					c.eventHandlers.onMessage(c, msg)
				}
			}
		}
	}
}

func (c *Client) HandleMessage(handler func(c *Client, msg *Message)) {
	c.eventHandlers.onMessage = handler
}
