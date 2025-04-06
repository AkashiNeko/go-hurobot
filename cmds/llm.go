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
		}{
			Prompt:     "你是一个群聊机器人，请你陪伴群友们聊天，注意请不要使用Markdown语法。",
			MaxHistory: 200,
			Enabled:    true,
			Debug:      false,
		}
		qbot.PsqlDB.Table("group_llm_configs").Create(map[string]any{
			"group_id":    msg.GroupID,
			"prompt":      llmConfig.Prompt,
			"max_history": llmConfig.MaxHistory,
			"enabled":     llmConfig.Enabled,
			"debug":       llmConfig.Debug,
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
				c.SendMsg(msg, "prompt 已更新")
			}
		}

	case "max-history":
		if args.Size == 2 {
			c.SendMsg(msg, fmt.Sprintf("max-history: %d", llmConfig.MaxHistory))
		} else {
			maxHistory, err := strconv.Atoi(args.Contents[2])
			if err != nil {
				c.SendMsg(msg, "请输入有效的数字")
				return
			}
			if maxHistory < 0 {
				c.SendMsg(msg, "max-history 不能为负值")
				return
			}
			if maxHistory > 300 {
				c.SendMsg(msg, "max-history 不能超过 300")
				return
			}
			err = qbot.PsqlDB.Table("group_llm_configs").
				Where("group_id = ?", msg.GroupID).
				Update("max_history", maxHistory).Error
			if err != nil {
				c.SendMsg(msg, "设置失败: "+err.Error())
			} else {
				c.SendMsg(msg, "max-history 已更新")
			}
		}

	case "enable":
		err := qbot.PsqlDB.Table("group_llm_configs").
			Where("group_id = ?", msg.GroupID).
			Update("enabled", true).Error
		if err != nil {
			c.SendMsg(msg, err.Error())
		} else {
			c.SendMsg(msg, "已启用本群 LLM 功能")
		}

	case "disable":
		err := qbot.PsqlDB.Table("group_llm_configs").
			Where("group_id = ?", msg.GroupID).
			Update("enabled", false).Error
		if err != nil {
			c.SendMsg(msg, err.Error())
		} else {
			c.SendMsg(msg, "已禁用本群 LLM 功能")
		}

	case "status":
		status := fmt.Sprintf("enabled: %v\nmax-history: %d\nprompt: %s",
			llmConfig.Enabled,
			llmConfig.MaxHistory,
			llmConfig.Prompt)
		c.SendMsg(msg, status)

	case "tokens":
		var user qbot.Users
		if args.Size == 2 {
			// 查看自己的 token 使用量
			err := qbot.PsqlDB.Where("user_id = ?", msg.UserID).First(&user).Error
			if err != nil {
				c.SendMsg(msg, "获取 token 使用量失败")
				return
			}
			c.SendMsg(msg, fmt.Sprintf("你的 token 使用量：%d", user.TokenUsage))
		} else if args.Size == 3 && args.Types[2] == qbot.At {
			// 查看其他人的 token 使用量
			targetID := str2uin64(args.Contents[2])
			err := qbot.PsqlDB.Where("user_id = ?", targetID).First(&user).Error
			if err != nil {
				c.SendMsg(msg, "获取 token 使用量失败")
				return
			}
			c.SendMsg(msg, fmt.Sprintf("用户 %s 的 token 使用量：%d", args.Contents[2], user.TokenUsage))
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

	default:
		c.SendMsg(msg, fmt.Sprintf("不能理解的参数 >>%s<<", args.Contents[1]))
	}
}
