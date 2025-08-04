package cmds

import (
	"fmt"
	"go-hurobot/config"
	"go-hurobot/qbot"
	"log"
	"os/exec"
	"strings"
	"time"
)

var workingDir string = "/tmp"

func truncateString(s string) string {
	s = encodeSpecialChars(s)
	const (
		maxLines    = 10
		maxChars    = 500
		truncateMsg = "\n输出过长，已自动截断"
	)

	lineCount := strings.Count(s, "\n") + 1

	if lineCount >= 11 {
		index := 0
		for range 10 {
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
	if msg.UserID != config.MasterID {
		return
	}

	if args.Size <= 1 {
		c.SendReplyMsg(msg, "Usage: sh <command>")
		return
	}

	rawcmd := decodeSpecialChars(msg.Raw[3:])

	if strings.HasPrefix(args.Contents[1], "cd") {
		absPath, err := exec.Command("bash", "-c",
			fmt.Sprintf("cd %s && %s && pwd", workingDir, rawcmd)).CombinedOutput()

		if err != nil {
			c.SendReplyMsg(msg, err.Error())
			return
		}

		workingDir = strings.TrimSpace(string(absPath))
		c.SendReplyMsg(msg, workingDir)
		return
	}

	cmd := exec.Command("bash", "-c", fmt.Sprintf("cd %s && %s", workingDir, rawcmd))

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
		if err == nil {
			// success
			c.SendReplyMsg(msg, truncateString(string(output)))
		} else {
			// failed
			c.SendReplyMsg(msg, fmt.Sprintf("%v\n%s", err, truncateString(string(output))))
		}
	case <-time.After(300 * time.Second):
		cmd.Process.Kill()
		c.SendReplyMsg(msg, fmt.Sprintf("Timeout: %q", rawcmd))
	}
}
