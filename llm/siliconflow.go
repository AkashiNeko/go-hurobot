package llm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go-hurobot/config"
	"io/ioutil"
	"net/http"
)

type LLMMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LLMRequest struct {
	Messages    []LLMMsg `json:"messages"`
	Model       string   `json:"model"`
	Stream      bool     `json:"stream"`
	Temperature float64  `json:"temperature"`
}

type LLMResponse struct {
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

func SendLLMRequest(supplier string, request *LLMRequest) (result *LLMResponse, err error) {
	var baseUrl, apiKey, erikaGrok2Key string
	switch supplier {
	case "grok":
		baseUrl = "https://grok.cclvi.cc/v1/chat/completions"
		apiKey = config.XaiApiKey
		erikaGrok2Key = config.ErikaGrok2Key
	case "siliconflow":
		baseUrl = "https://api.siliconflow.com/v1/chat/completions"
		apiKey = config.SiliconflowApiKey
	default:
		return nil, errors.New("invalid supplier")
	}

	switch {
	case request == nil:
		return nil, errors.New("request is nil")
	}

	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	requestJson := string(jsonBytes)

	client := &http.Client{}

	// Use custom proxy
	if config.ProxyURL.Host != "" {
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(config.ProxyURL),
		}
	}

	req, err := http.NewRequest("POST", baseUrl, bytes.NewBuffer([]byte(requestJson)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	if erikaGrok2Key != "" {
		req.Header.Set("X-Proxy-Key", erikaGrok2Key)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		ret := &LLMResponse{}
		if err := json.Unmarshal(body, ret); err != nil {
			return nil, err
		}
		return ret, nil
	}

	return nil, fmt.Errorf("%s\n\n%s", resp.Status, string(body))
}
