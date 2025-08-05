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

func (c *Client) SetGroupBan(groupID uint64, userID uint64, duration int) error {
	req := cqRequest{
		Action: "set_group_ban",
		Params: map[string]any{
			"group_id": groupID,
			"user_id":  userID,
			"duration": duration,
		},
	}
	err := c.sendJson(&req)
	return err
}

func (c *Client) SetGroupEssence(msgID uint64) error {
	req := cqRequest{
		Action: "set_essence_msg",
		Params: map[string]any{
			"message_id": msgID,
		},
	}
	err := c.sendJson(&req)
	return err
}

func (c *Client) DeleteGroupEssence(msgID uint64) error {
	req := cqRequest{
		Action: "delete_essence_msg",
		Params: map[string]any{
			"message_id": msgID,
		},
	}
	err := c.sendJson(&req)
	return err
}

func (c *Client) DeleteMsg(msgID uint64) error {
	req := cqRequest{
		Action: "delete_msg",
		Params: map[string]any{
			"message_id": msgID,
		},
	}
	err := c.sendJson(&req)
	return err
}

func (c *Client) SendRecord(msg *Message, file string) {
	c.SendMsg(msg, CQRecord(file))
}

func (c *Client) SendReplyMsg(msg *Message, message string) {
	c.SendMsg(msg, CQReply(msg.MsgID)+message)
}

func (c *Client) SendMsg(msg *Message, message string) {
	if msg.GroupID == 0 {
		c.SendPrivateMsg(msg.UserID, message, false)
	} else {
		c.SendGroupMsg(msg.GroupID, message, false)
	}
}

func (c *Client) SendImage(msg *Message, url string) {
	c.SendMsg(msg, CQImage(url))
}

func (c *Client) GetGroupMemberInfo(groupID uint64, userID uint64, noCache bool) (*GroupMemberInfo, error) {
	req := cqRequest{
		Action: "get_group_member_info",
		Params: map[string]any{
			"group_id": groupID,
			"user_id":  userID,
			"no_cache": noCache,
		},
	}
	resp, err := c.sendJsonWithEcho(&req)
	if err != nil {
		return nil, err
	}
	return &resp.Data.GroupMemberInfo, nil
}
