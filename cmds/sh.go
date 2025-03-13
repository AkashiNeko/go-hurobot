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

func cmd_sh(c *qbot.Client, msg *qbot.Message, args *ArgsList) {
	if msg.UserID != config.MasterID {
		c.SendReplyMsg(msg, fmt.Sprintf("%s: Permission denied", args.Contents[0]))
		return
	}
	if args.Size <= 1 {
		c.SendReplyMsg(msg, "Usage: sh <linux command>")
		return
	}

	if strings.HasPrefix(args.Contents[1], "cd") {
		if args.Size > 2 {
			absPath, err := exec.Command("realpath", strings.TrimSpace(workingDir)).Output()
			if err != nil {
				c.SendReplyMsg(msg, err.Error())
				return
			}
			workingDir = string(absPath)
		} else {
			workingDir = os.Getenv("HOME")
		}
		c.SendReplyMsg(msg, workingDir)
		return
	}

	cmd := exec.Command("bash", "-c", strings.Join(args.Contents[1:], " "))
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	done := make(chan error, 1)
	var output []byte

	go func() {
		var err error
		output, err = cmd.CombinedOutput()
		log.Printf("run command: %s, output: %s, error: %v",
			strings.Join(args.Contents[1:], " "), string(output), err)
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			c.SendReplyMsg(msg, fmt.Sprintf("%v\n%s", err, string(output)))
			return
		}
		c.SendReplyMsg(msg, string(output))
	case <-time.After(30 * time.Second):
		cmd.Process.Kill()
		c.SendReplyMsg(msg, fmt.Sprintf("命令执行超时: %s",
			strings.Join(args.Contents[1:], " ")))
	}
}
