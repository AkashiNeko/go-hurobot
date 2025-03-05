package qbot

import (
	"encoding/json"
)

func (c *Client) handleEvents(postType *string, msgStr *[]byte, jsonMap *map[string]any) {
	switch *postType {
	case "meta_event":
		// heartbeat, connection state..
	case "notice":
		// TODO
	case "message":
		switch (*jsonMap)["message_type"] {
		case "private":
			fallthrough
		case "group":
			if c.eventHandlers.onMessage != nil {
				var msg Message
				if json.Unmarshal(*msgStr, &msg) == nil {
					go c.eventHandlers.onMessage(c, &msg)
				}
			}
		}
	}
}

func (c *Client) HandleMessage(handler func(c *Client, msg *Message)) {
	c.eventHandlers.onMessage = handler
}
