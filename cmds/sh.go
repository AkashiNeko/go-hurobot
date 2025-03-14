package cmds

import (
	"fmt"
	"go-hurobot/qbot"
	"log"
	"os/exec"
	"strings"
	"time"
)

var workingDir string = "/home/qwq"

func truncateString(s string) string {
	const (
		maxLines    = 10
		maxChars    = 500
		truncateMsg = "\n输出过长，已自动截断"
	)

	lineCount := strings.Count(s, "\n") + 1

	if lineCount >= 11 {
		index := 0
		for i := 0; i < 10; i++ {
			index = strings.Index(s[index:], "\n") + 1 + index
			if index == 0 {
				return s
			}
		}
		return s[:index] + truncateMsg
	}

	if len(s) > maxChars {
		return s[:maxChars] + truncateMsg
	}

	return s
}

func cmd_sh(c *qbot.Client, msg *qbot.Message, args *ArgsList) {
	if msg.GroupID == 0 {
		c.SendReplyMsg(msg, "请在群内使用")
		return
	}

	if args.Size <= 1 {
		c.SendReplyMsg(msg, "Usage: sh <linux command>")
		return
	}

	rawcmd := decodeSpecialChars(msg.Raw[3:])

	if strings.HasPrefix(args.Contents[1], "cd") {

		absPath, err := exec.Command("docker", "exec", "-i", "-u", "1000",
			"-w", workingDir, "ubuntu", "bash", "-c",
			fmt.Sprintf("%s && pwd", rawcmd)).CombinedOutput()

		if err != nil {
			c.SendReplyMsg(msg, err.Error())
			return
		}

		workingDir = strings.TrimSpace(string(absPath))
		c.SendReplyMsg(msg, workingDir)
		return
	}

	cmd := exec.Command("docker", "exec", "-i", "-u", "1000",
		"-w", workingDir, "ubuntu", "bash", "-c", rawcmd)

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
		c.SendReplyMsg(msg, truncateString(string(output)))
	case <-time.After(300 * time.Second):
		cmd.Process.Kill()
		c.SendReplyMsg(msg, fmt.Sprintf("命令执行超时: %s",
			strings.Join(args.Contents[1:], " ")))
	}
}
