package cmds

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go-hurobot/qbot"
)

type TTSVoice struct {
	Name  string `gorm:"column:name"`
	Local string `gorm:"column:local"`
}

func cmd_tts(c *qbot.Client, msg *qbot.Message, args *ArgsList) {
	const help = "Usage:\n  tts search <voice_name>\n  tts <voice_name> <text>"
	for _, t := range args.Types {
		if t != qbot.Text {
			c.SendMsg(msg, help)
			return
		}
	}

	if args.Size < 2 {
		c.SendMsg(msg, help)
		return
	}

	switch args.Contents[1] {
	case "help", "-h", "-?", "--help", "?":
		c.SendMsg(msg, help)
		return
	case "search", "q":
		if args.Size < 3 {
			c.SendMsg(msg, "tts search <voice_name>")
			return
		}
		keyword := strings.Join(args.Contents[2:], " ")
		var voices []TTSVoice
		result := qbot.PsqlDB.Table("tts_voices").
			Where("name ~ ?", keyword).
			Limit(11).
			Find(&voices)

		if result.Error != nil {
			c.SendMsg(msg, result.Error.Error())
			return
		}

		if len(voices) == 0 {
			c.SendMsg(msg, "[]")
			return
		}

		var output strings.Builder
		for i, voice := range voices {
			if i == 10 {
				output.WriteString("...\n")
				break
			}
			output.WriteString(fmt.Sprintf("%s (%s)\n", voice.Name, voice.Local))
		}
		c.SendMsg(msg, output.String())

	default:
		if args.Size < 3 {
			c.SendMsg(msg, "tts "+args.Contents[1]+" <text>")
			return
		}

		voiceName := args.Contents[1]
		text := strings.Join(args.Contents[2:], " ")

		var voice TTSVoice
		result := qbot.PsqlDB.Table("tts_voices").
			Where("name = ?", voiceName).
			First(&voice)

		if result.Error != nil {
			c.SendMsg(msg, "voice not found: "+voiceName)
			return
		}

		cmd := exec.Command("./tts", voiceName, text)
		output, err := cmd.CombinedOutput()

		if err != nil {
			c.SendMsg(msg, err.Error())
			return
		}

		if cmd.ProcessState.ExitCode() != 0 {
			c.SendMsg(msg, string(output))
			return
		}

		filename := strings.TrimSpace(string(output))
		filePath := filepath.Join("out", filename)

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			c.SendMsg(msg, "failed: file not found")
			return
		}
		if fileInfo.Size() == 0 {
			c.SendMsg(msg, "failed: file is empty")
			return
		}

		c.SendRecord(msg, filePath)
	}
}
