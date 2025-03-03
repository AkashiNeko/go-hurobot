package qbot

import (
	"encoding/json"
)

func (c *Client) handleEvents(postType *string, msg *[]byte, jsonMap *map[string]any) {
	switch *postType {
	case "meta_event":
		// heartbeat, connection state..
	case "notice":
		// TODO
	case "message":
		switch (*jsonMap)["message_type"] {
		case "private":
			// TODO
			if c.eventHandlers.onPrivateMessage != nil {
				var privateMessage PrivateMessage
				if json.Unmarshal(*msg, &privateMessage) == nil {
					go c.eventHandlers.onPrivateMessage(c, privateMessage)
				}
			}
		case "group":
			// TODO
			if c.eventHandlers.onGroupMessage != nil {
				var groupMessage GroupMessage
				if json.Unmarshal(*msg, &groupMessage) == nil {
					go c.eventHandlers.onGroupMessage(c, groupMessage)
				}
			}
		}
	}
}

func (c *Client) HandleGroupMessage(handler func(c *Client, msg GroupMessage)) {
	c.eventHandlers.onGroupMessage = handler
}

func (c *Client) HandlePrivateMessage(handler func(c *Client, msg PrivateMessage)) {
	c.eventHandlers.onPrivateMessage = handler
}
