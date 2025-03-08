package cmds

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go-hurobot/config"
	"go-hurobot/qbot"
	"io/ioutil"
	"net/http"
	"strconv"
)

func sendRequest(requestJson string) (result string, err error) {
	apiKey := config.XaiApiKey
	if apiKey == "" {
		return "", errors.New("No x.ai api key")
	}

	client := &http.Client{}

	// Use custom proxy
	if config.ProxyURL.Host != "" {
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(&config.ProxyURL),
		}
	}

	req, err := http.NewRequest("POST", "https://api.x.ai/v1/chat/completions", bytes.NewBuffer([]byte(requestJson)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode == http.StatusOK {
		return string(body), nil
	}

	return "", errors.New(fmt.Sprintf("%s\n\n%s", resp.Status, string(body)))
}

func makeGrokRequest(args []string) string {
	if len(args) == 0 {
		return "请输入文本"
	}
	type Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type Request struct {
		Messages    []Message `json:"messages"`
		Model       string    `json:"model"`
		Stream      bool      `json:"stream"`
		Temperature float64   `json:"temperature"`
	}
	request := &Request{}
	request.Model = "grok-2-latest"
	prevArg := ""
	for _, arg := range args {
		switch prevArg {
		case "-s":
			request.Messages = append(request.Messages, Message{
				Role:    "system",
				Content: arg,
			})
			prevArg = ""
		case "-a":
			request.Messages = append(request.Messages, Message{
				Role:    "assistant",
				Content: arg,
			})
			prevArg = ""
		case "-u":
			request.Messages = append(request.Messages, Message{
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
	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return "啊哦！出错了！@氟氟"
	}
	ret, err := sendRequest(string(jsonBytes))
	if err != nil {
		return err.Error()
	}
	type Response struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		Model   string `json:"model"`
		Choices []struct {
			Index   int `json:"index"`
			Message struct {
				Role    string      `json:"role"`
				Content string      `json:"content"`
				Refusal interface{} `json:"refusal"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens        int `json:"prompt_tokens"`
			CompletionTokens    int `json:"completion_tokens"`
			ReasoningTokens     int `json:"reasoning_tokens"`
			TotalTokens         int `json:"total_tokens"`
			PromptTokensDetails struct {
				TextTokens   int `json:"text_tokens"`
				AudioTokens  int `json:"audio_tokens"`
				ImageTokens  int `json:"image_tokens"`
				CachedTokens int `json:"cached_tokens"`
			} `json:"prompt_tokens_details"`
		} `json:"usage"`
		SystemFingerprint string `json:"system_fingerprint"`
	}
	jsonRet := &Response{}
	json.Unmarshal([]byte(ret), jsonRet)
	return fmt.Sprintf("%s\n\nmodel: %s\nprompt_tokens: %d\ncompletion_tokens: %d\ncreated: %d\nid: %s",
		jsonRet.Choices[0].Message.Content,
		jsonRet.Model,
		jsonRet.Usage.PromptTokens,
		jsonRet.Usage.CompletionTokens,
		jsonRet.Created,
		jsonRet.ID,
	)
}

func grok2(c *qbot.Client, args []string, raw *qbot.Message) {
	const help = "Usage: grok2 <option> [-s <system content>] [-a <assistant content>] [-u <user content>] [-m <model>] [-t <temperature>]"
	if len(args) == 1 {
		c.SendReplyMsg(raw, help)
		return
	}
	switch args[1] {
	case "debug":
		if len(args) < 4 {
			c.SendReplyMsg(raw, help)
			return
		}
		c.SendReplyMsg(raw, makeGrokRequest(args[2:]))
	case "help":
		c.SendReplyMsg(raw, help)
	default:
		c.SendReplyMsg(raw, fmt.Sprintf("不能理解参数：grok2 >>%s<<\n发送 grok2 help 查看帮助", args[1]))
	}
}
