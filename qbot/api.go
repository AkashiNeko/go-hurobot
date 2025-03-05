package qbot

func (c *Client) SendPrivateMsg(userID uint64, message string, autoEscape bool) (uint64, error) {
	req := cqRequest{
		Action: "send_private_msg",
		Params: map[string]any{
			"user_id": userID,
			// "group_id":    groupID,
			"message":     message,
			"auto_escape": autoEscape,
		},
	}
	resp, err := c.sendJSON(&req)
	if err != nil {
		return 0, err
	}
	return resp.Data.MessageId, nil
}

func (c *Client) SendGroupMsg(groupID uint64, message string, autoEscape bool) (uint64, error) {
	req := cqRequest{
		Action: "send_group_msg",
		Params: map[string]any{
			"group_id":    groupID,
			"message":     message,
			"auto_escape": autoEscape,
		},
	}

	resp, err := c.sendJSON(&req)
	if err != nil {
		return 0, err
	}
	return resp.Data.MessageId, nil
}

func (c *Client) SendReplyMsg(raw *Message, message string) {
	if raw.GroupID == 0 {
		c.SendPrivateMsg(raw.Sender.UserID, message, false)
	} else {
		c.SendGroupMsg(raw.GroupID, message, false)
	}
}
