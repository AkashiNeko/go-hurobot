package cmds

import (
	"fmt"
	"strconv"
	"strings"

	"go-hurobot/qbot"
)

func cmd_llm(c *qbot.Client, msg *qbot.Message, args *ArgsList) {
	if args.Size < 2 {
		c.SendMsg(msg, "Usage:\nllm prompt [新提示词]\nllm max-history [能看见的历史消息数]\nllm enable/disable\nllm status")
		return
	}

	var llmConfig struct {
		Prompt     string
		MaxHistory int
		Enabled    bool
		Debug      bool
		Supplier   string
		Model      string
	}

	err := qbot.PsqlDB.Table("group_llm_configs").
		Where("group_id = ?", msg.GroupID).
		First(&llmConfig).Error

	if err != nil {
		llmConfig = struct {
			Prompt     string
			MaxHistory int
			Enabled    bool
			Debug      bool
			Supplier   string
			Model      string
		}{
			Prompt:     "你是一个群聊机器人，请你陪伴群友们聊天，注意请不要使用Markdown语法。",
			MaxHistory: 200,
			Enabled:    true,
			Debug:      false,
			Supplier:   "siliconflow",
			Model:      "deepseek-ai/DeepSeek-V2.5",
		}
		qbot.PsqlDB.Table("group_llm_configs").Create(map[string]any{
			"group_id":    msg.GroupID,
			"prompt":      llmConfig.Prompt,
			"max_history": llmConfig.MaxHistory,
			"enabled":     llmConfig.Enabled,
			"debug":       llmConfig.Debug,
			"supplier":    "siliconflow",
			"model":       "deepseek-ai/DeepSeek-V2.5",
		})
	}

	switch args.Contents[1] {
	case "prompt":
		if args.Size == 2 {
			c.SendMsg(msg, fmt.Sprintf("prompt: %s", llmConfig.Prompt))
		} else {
			newPrompt := strings.Join(args.Contents[2:], " ")
			err := qbot.PsqlDB.Table("group_llm_configs").
				Where("group_id = ?", msg.GroupID).
				Update("prompt", newPrompt).Error
			if err != nil {
				c.SendMsg(msg, err.Error())
			} else {
				c.SendMsg(msg, "prompt updated")
			}
		}

	case "max-history":
		if args.Size == 2 {
			c.SendMsg(msg, fmt.Sprintf("max-history: %d", llmConfig.MaxHistory))
		} else {
			maxHistory, err := strconv.Atoi(args.Contents[2])
			if err != nil {
				c.SendMsg(msg, "Enter a valid number")
				return
			}
			if maxHistory < 0 {
				c.SendMsg(msg, "max-history cannot be negative")
				return
			}
			if maxHistory > 300 {
				c.SendMsg(msg, "max-history cannot exceed 300")
				return
			}
			err = qbot.PsqlDB.Table("group_llm_configs").
				Where("group_id = ?", msg.GroupID).
				Update("max_history", maxHistory).Error
			if err != nil {
				c.SendMsg(msg, "Failed: "+err.Error())
			} else {
				c.SendMsg(msg, "max-history updated")
			}
		}

	case "enable":
		err := qbot.PsqlDB.Table("group_llm_configs").
			Where("group_id = ?", msg.GroupID).
			Update("enabled", true).Error
		if err != nil {
			c.SendMsg(msg, err.Error())
		} else {
			c.SendMsg(msg, "Enabled LLM")
		}

	case "disable":
		err := qbot.PsqlDB.Table("group_llm_configs").
			Where("group_id = ?", msg.GroupID).
			Update("enabled", false).Error
		if err != nil {
			c.SendMsg(msg, err.Error())
		} else {
			c.SendMsg(msg, "Disabled LLM")
		}

	case "status":
		status := fmt.Sprintf("enabled: %v\nmax-history: %d\nsupplier: %q\nmodel: %q\nprompt: %q",
			llmConfig.Enabled,
			llmConfig.MaxHistory,
			llmConfig.Supplier,
			llmConfig.Model,
			llmConfig.Prompt,
		)
		c.SendMsg(msg, status)

	case "tokens":
		var user qbot.Users
		if args.Size == 2 {
			err := qbot.PsqlDB.Where("user_id = ?", msg.UserID).First(&user).Error
			if err != nil {
				c.SendMsg(msg, "Failed to get token usage")
				return
			}
			c.SendMsg(msg, fmt.Sprintf("Token usage: %d", user.TokenUsage))
		} else if args.Size == 3 && args.Types[2] == qbot.At {
			targetID := str2uin64(args.Contents[2])
			err := qbot.PsqlDB.Where("user_id = ?", targetID).First(&user).Error
			if err != nil {
				c.SendMsg(msg, "Failed to get token usage")
				return
			}
			c.SendMsg(msg, fmt.Sprintf("Token usage for %s: %d", args.Contents[2], user.TokenUsage))
		} else {
			c.SendMsg(msg, "Usage:\nllm tokens\nllm tokens @user")
		}

	case "debug":
		if args.Size == 2 {
			c.SendMsg(msg, fmt.Sprintf("debug: %v", llmConfig.Debug))
		} else {
			debugValue := strings.ToLower(args.Contents[2])
			if debugValue != "on" && debugValue != "off" {
				return
			}
			newDebug := debugValue == "on"
			err := qbot.PsqlDB.Table("group_llm_configs").
				Where("group_id = ?", msg.GroupID).
				Update("debug", newDebug).Error
			if err != nil {
				c.SendMsg(msg, err.Error())
			} else {
				c.SendMsg(msg, fmt.Sprintf("debug = %v", newDebug))
			}
		}

	case "model":
		if args.Size == 2 {
			c.SendMsg(msg, fmt.Sprintf("model: %s", llmConfig.Model))
		} else {
			newModel := args.Contents[2]
			err := qbot.PsqlDB.Table("group_llm_configs").
				Where("group_id = ?", msg.GroupID).
				Update("model", newModel).Error
			if err != nil {
				c.SendMsg(msg, err.Error())
			} else {
				c.SendMsg(msg, fmt.Sprintf("model updated to %s", newModel))
			}
		}

	default:
		c.SendMsg(msg, fmt.Sprintf("Unrecognized parameter >>%s<<", args.Contents[1]))
	}
}
