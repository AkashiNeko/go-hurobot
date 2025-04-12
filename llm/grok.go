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

type Grok2Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Grok2Request struct {
	Messages    []Grok2Message `json:"messages"`
	Model       string         `json:"model"`
	Stream      bool           `json:"stream"`
	Temperature float64        `json:"temperature"`
}

type Grok2Response struct {
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

func SendGrok2Request(request *Grok2Request) (result *Grok2Response, err error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	requestJson := string(jsonBytes)

	apiKey := config.SiliconflowApiKey
	if apiKey == "" {
		return nil, errors.New("no x.ai api key")
	}

	client := &http.Client{}

	// Use custom proxy
	if config.ProxyURL.Host != "" {
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(config.ProxyURL),
		}
	}

	requestURL := "https://api.x.ai/v1/chat/completions"
	if config.ErikaGrok2Key != "" {
		requestURL = "https://grok.cclvi.cc/v1/chat/completions"
	}

	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer([]byte(requestJson)))
	if err != nil {
		return nil, err
	}

	if config.ErikaGrok2Key != "" {
		req.Header.Set("X-Proxy-Key", config.ErikaGrok2Key)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

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
		ret := &Grok2Response{}
		if err := json.Unmarshal(body, ret); err != nil {
			return nil, err
		}
		return ret, nil
	}

	return nil, fmt.Errorf("%s\n\n%s", resp.Status, string(body))
}
