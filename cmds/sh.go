package cmds

import (
	"fmt"
	"go-hurobot/config"
	"go-hurobot/qbot"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

var workingDir string = ""

func cmd_sh(c *qbot.Client, args []string, raw *qbot.Message) {
	if raw.Sender.UserID != config.MasterID {
		c.SendReplyMsg(raw, fmt.Sprintf("%s: Permission denied", args[0]))
		return
	}
	if len(args) <= 1 {
		c.SendReplyMsg(raw, "Usage: sh <linux command>")
		return
	}

	if strings.HasPrefix(args[1], "cd") {
		if len(args) > 2 {
			absPath, err := exec.Command("realpath", strings.TrimSpace(workingDir)).Output()
			if err != nil {
				c.SendReplyMsg(raw, err.Error())
				return
			}
			workingDir = string(absPath)
		} else {
			workingDir = os.Getenv("HOME")
		}
		c.SendReplyMsg(raw, workingDir)
		return
	}

	cmd := exec.Command("bash", "-c", strings.Join(args[1:], " "))
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	done := make(chan error, 1)
	var output []byte

	go func() {
		var err error
		output, err = cmd.CombinedOutput()
		log.Printf("run command: %s, output: %s, error: %v",
			strings.Join(args[1:], " "), string(output), err)
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			c.SendReplyMsg(raw, fmt.Sprintf("%v\n%s", err, string(output)))
			return
		}
		c.SendReplyMsg(raw, string(output))
	case <-time.After(30 * time.Second):
		cmd.Process.Kill()
		c.SendReplyMsg(raw, fmt.Sprintf("命令执行超时: %s",
			strings.Join(args[1:], " ")))
	}
}
