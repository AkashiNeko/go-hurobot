package cmds

import (
	"fmt"
	"go-hurobot/qbot"
	"os/exec"
	"strings"
)

func cmd_info(c *qbot.Client, msg *qbot.Message, args *ArgsList) {
	cmd := exec.Command("top", "-l", "1", "-n", "0")
	output, err := cmd.Output()
	if err != nil {
		c.SendReplyMsg(msg, fmt.Sprintf("Failed to get system info: %v", err))
		return
	}

	c.SendReplyMsg(msg, strings.TrimSpace(string(output)))
}
