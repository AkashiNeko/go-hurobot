package cmds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-hurobot/config"
	"go-hurobot/qbot"
)

type ImageGenerationRequest struct {
	Model             string  `json:"model"`
	Prompt            string  `json:"prompt"`
	NegativePrompt    string  `json:"negative_prompt,omitempty"`
	ImageSize         string  `json:"image_size"`
	BatchSize         int     `json:"batch_size"`
	Seed              *int64  `json:"seed,omitempty"`
	NumInferenceSteps int     `json:"num_inference_steps"`
	GuidanceScale     float64 `json:"guidance_scale"`
}

type ImageGenerationResponse struct {
	Images []struct {
		URL string `json:"url"`
	} `json:"images"`
	Timings struct {
		Inference float64 `json:"inference"`
	} `json:"timings"`
	Seed int64 `json:"seed"`
}

func cmd_draw(c *qbot.Client, msg *qbot.Message, args *ArgsList) {
	if args.Size < 2 {
		helpMsg := `Usage: draw <prompt> [options]

Options:
  --negative <负面提示词>
  --size <尺寸> (1024x1024、1024x2048、1536x1024、1536x2048、2048x1152、1152x2048)
  --steps <步数> (default 20)
  --scale <引导强度> (default 7.5)
  --seed <种子> (default random)`
		c.SendMsg(msg, helpMsg)
		return
	}

	if config.ApiKey == "" {
		c.SendMsg(msg, "No API key")
		return
	}

	prompt, negativePrompt, imageSize, steps, scale, seed := parseDrawArgs(args.Contents[1:])

	if prompt == "" {
		c.SendMsg(msg, "Please provide a prompt")
		return
	}

	c.SendMsg(msg, "Image generating...")

	reqData := ImageGenerationRequest{
		Model:             "Kwai-Kolors/Kolors",
		Prompt:            prompt,
		NegativePrompt:    negativePrompt,
		ImageSize:         imageSize,
		BatchSize:         1,
		NumInferenceSteps: steps,
		GuidanceScale:     scale,
	}

	if seed != nil {
		reqData.Seed = seed
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		c.SendMsg(msg, fmt.Sprintf("%v", err))
		return
	}

	req, err := http.NewRequest("POST", "https://api.siliconflow.cn/v1/images/generations", bytes.NewBuffer(jsonData))
	if err != nil {
		c.SendMsg(msg, fmt.Sprintf("%v", err))
		return
	}

	req.Header.Set("Authorization", "Bearer "+config.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		c.SendMsg(msg, fmt.Sprintf("%v", err))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.SendMsg(msg, fmt.Sprintf("%v", err))
		return
	}

	if resp.StatusCode != 200 {
		c.SendMsg(msg, fmt.Sprintf("%d\n%s", resp.StatusCode, string(body)))
		return
	}

	var imgResp ImageGenerationResponse
	if err := json.Unmarshal(body, &imgResp); err != nil {
		c.SendMsg(msg, fmt.Sprintf("%v", err))
		return
	}

	if len(imgResp.Images) == 0 {
		c.SendMsg(msg, "错误: 未生成任何图片")
		return
	}

	inferenceTimeMs := int(imgResp.Timings.Inference * 1000)
	infoMsg := fmt.Sprintf("timings: %dms\nseed: %d\nscale: %f",
		inferenceTimeMs, imgResp.Seed, scale)

	imageURL := imgResp.Images[0].URL
	c.SendMsg(msg, qbot.CQReply(msg.MsgID)+qbot.CQImage(imageURL)+infoMsg)
}

func parseDrawArgs(args []string) (prompt, negativePrompt, imageSize string, steps int, scale float64, seed *int64) {

	imageSize = "1024x1024"
	steps = 20
	scale = 7.5

	var promptParts []string
	i := 0

	for i < len(args) {
		arg := args[i]

		switch arg {
		case "--negative":
			if i+1 < len(args) {
				negativePrompt = args[i+1]
				i += 2
			} else {
				i++
			}
		case "--size":
			if i+1 < len(args) {
				size := args[i+1]
				if isValidSize(size) {
					imageSize = size
				}
				i += 2
			} else {
				i++
			}
		case "--steps":
			if i+1 < len(args) {
				if s, err := strconv.Atoi(args[i+1]); err == nil && s >= 1 && s <= 50 {
					steps = s
				}
				i += 2
			} else {
				i++
			}
		case "--scale":
			if i+1 < len(args) {
				if s, err := strconv.ParseFloat(args[i+1], 64); err == nil && s >= 1.0 && s <= 20.0 {
					scale = s
				}
				i += 2
			} else {
				i++
			}
		case "--seed":
			if i+1 < len(args) {
				if s, err := strconv.ParseInt(args[i+1], 10, 64); err == nil {
					seed = &s
				}
				i += 2
			} else {
				i++
			}
		default:
			promptParts = append(promptParts, arg)
			i++
		}
	}

	prompt = strings.Join(promptParts, " ")
	return
}

func isValidSize(size string) bool {
	validSizes := []string{"512x512", "768x768", "1024x1024", "1024x768", "768x1024"}
	for _, valid := range validSizes {
		if size == valid {
			return true
		}
	}
	return false
}
