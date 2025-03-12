package qbot

func (c *Client) SendPrivateMsg(userID uint64, message string, autoEscape bool) (uint64, error) {
	if message == "" {
		message = " "
	}
	req := cqRequest{
		Action: "send_private_msg",
		Params: map[string]any{
			"user_id":     userID,
			"message":     message,
			"auto_escape": autoEscape,
		},
	}
	resp, err := c.sendJsonWithEcho(&req)
	if err != nil {
		return 0, err
	}
	return resp.Data.MessageId, nil
}

func (c *Client) SendGroupMsg(groupID uint64, message string, autoEscape bool) (uint64, error) {
	if message == "" {
		message = " "
	}
	req := cqRequest{
		Action: "send_group_msg",
		Params: map[string]any{
			"group_id":    groupID,
			"message":     message,
			"auto_escape": autoEscape,
		},
	}

	resp, err := c.sendJsonWithEcho(&req)
	if err != nil {
		return 0, err
	}
	return resp.Data.MessageId, nil
}

func (c *Client) SetGroupSpecialTitle(groupID uint64, userID uint64, specialTitle string) error {
	req := cqRequest{
		Action: "set_group_special_title",
		Params: map[string]any{
			"group_id":      groupID,
			"user_id":       userID,
			"special_title": specialTitle,
		},
	}
	err := c.sendJson(&req)
	return err
}

func (c *Client) SetGroupName(groupID uint64, groupName string) error {
	req := cqRequest{
		Action: "set_group_name",
		Params: map[string]any{
			"group_id":   groupID,
			"group_name": groupName,
		},
	}
	err := c.sendJson(&req)
	return err
}

func (c *Client) SetGroupAdmin(groupID uint64, userID uint64, enable bool) error {
	req := cqRequest{
		Action: "set_group_admin",
		Params: map[string]any{
			"group_id": groupID,
			"user_id":  userID,
			"enable":   enable,
		},
	}
	err := c.sendJson(&req)
	return err
}

func (c *Client) SendReplyMsg(raw *Message, message string) {
	if raw.GroupID == 0 {
		c.SendPrivateMsg(raw.Sender.UserID, message, false)
	} else {
		c.SendGroupMsg(raw.GroupID, message, false)
	}
}
