package cmds

import (
	"fmt"
	"go-hurobot/llm"
	"go-hurobot/qbot"
	"strconv"
)

func callGrok2(args []string) string {
	if len(args) == 0 {
		return "请输入文本"
	}

	request := &llm.Grok2Request{}
	request.Model = "grok-2-latest"
	prevArg := ""
	for _, arg := range args {
		switch prevArg {
		case "-s":
			request.Messages = append(request.Messages, llm.Grok2Message{
				Role:    "system",
				Content: arg,
			})
			prevArg = ""
		case "-a":
			request.Messages = append(request.Messages, llm.Grok2Message{
				Role:    "assistant",
				Content: arg,
			})
			prevArg = ""
		case "-u":
			request.Messages = append(request.Messages, llm.Grok2Message{
				Role:    "user",
				Content: arg,
			})
			prevArg = ""
		case "-m":
			request.Model = arg
			prevArg = ""
		case "-t":
			if t, err := strconv.ParseFloat(arg, 64); err == nil {
				request.Temperature = t
			} else {
				return fmt.Sprintf("temperature 必须是一个数字：... -t >>%s<<", arg)
			}
			prevArg = ""
		case "":
			prevArg = arg
		default:
			return fmt.Sprintf("不能理解参数：... >>%s<<", prevArg)
		}
	}
	if prevArg != "" {
		return fmt.Sprintf("不完整的参数：... >>%s<<", prevArg)
	}
	ret, err := llm.SendGrok2Request(request)
	if err != nil {
		return err.Error()
	}

	return fmt.Sprintf("%s\n\nmodel: %s\nprompt_tokens: %d\ncompletion_tokens: %d\ncreated: %d\nid: %s",
		ret.Choices[0].Message.Content,
		ret.Model,
		ret.Usage.PromptTokens,
		ret.Usage.CompletionTokens,
		ret.Created,
		ret.ID,
	)
}

func cmd_grok2(c *qbot.Client, raw *qbot.Message, args *ArgsList) {
	const help = "Usage: grok2 <option> [-s <system content>] [-a <assistant content>] [-u <user content>] [-m <model>] [-t <temperature>]"
	if args.Size == 1 {
		c.SendReplyMsg(raw, help)
		return
	}
	switch args.Contents[1] {
	case "debug":
		if args.Size < 4 {
			c.SendReplyMsg(raw, help)
			return
		}
		c.SendReplyMsg(raw, callGrok2(args.Contents[2:]))
	case "help":
		c.SendReplyMsg(raw, help)
	default:
		c.SendReplyMsg(raw, fmt.Sprintf("不能理解参数：grok2 >>%s<<\n发送 grok2 help 查看帮助", args.Contents[1]))
	}
}
