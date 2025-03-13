package qbot

import (
	"encoding/json"
)

func ParseMsgJson(raw *messageJson) *Message {
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
		var jsonData map[string]string
		err := json.Unmarshal([]byte(msg.Data), &jsonData)
		if err != nil {
			return nil
		}
		switch msg.Type {
		case "text":
			result.Array = append(result.Array, MsgItem{
				Type:    Text,
				Content: jsonData["text"],
			})
		case "image":
			result.Array = append(result.Array, MsgItem{
				Type:    Image,
				Content: jsonData["url"],
			})
		case "record":
			result.Array = append(result.Array, MsgItem{
				Type:    Record,
				Content: jsonData["path"],
			})
		case "at":
			result.Array = append(result.Array, MsgItem{
				Type:    At,
				Content: jsonData["id"],
			})
		case "reply":
			result.Array = append(result.Array, MsgItem{
				Type:    Reply,
				Content: jsonData["id"],
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
				if msg := ParseMsgJson(msgJson); msg != nil {
					go saveDatabase(msg)
					go c.eventHandlers.onMessage(c, msg)
				}
			}
		}
	}
}

func (c *Client) HandleMessage(handler func(c *Client, msg *Message)) {
	c.eventHandlers.onMessage = handler
}
