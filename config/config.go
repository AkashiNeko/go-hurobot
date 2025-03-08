package config

import (
	"log"
	"net/url"
	"os"
	"strconv"
)

var (
	// export
	XaiApiKey   string
	NapcatWSURL string
	MasterID    uint64
	ProxyURL    url.URL
)

const (
	// environment values
	env_NAPCAT_HOST  = "NAPCAT_HOST"
	env_ACCESS_TOKEN = "ACCESS_TOKEN"
	env_XAI_API_KEY  = "XAI_API_KEY"
	env_MASTER_ID    = "MASTER_ID"
	env_PROXY_URL    = "PROXY_URL"
)

func init() {
	napcatHost := os.Getenv(env_NAPCAT_HOST)
	if napcatHost == "" {
		napcatHost = "127.0.0.1:3001"
	}
	accessToken := os.Getenv(env_ACCESS_TOKEN)
	XaiApiKey = os.Getenv(env_XAI_API_KEY)

	NapcatWSURL = "ws://" + napcatHost
	if accessToken != "" {
		NapcatWSURL += "?access_token=" + accessToken
	}

	masterIdStr := os.Getenv(env_MASTER_ID)
	if masterIdStr == "" {
		masterIdStr = "1006554341"
	}
	adminId, err := strconv.Atoi(masterIdStr)
	if err != nil {
		log.Fatalf("Parse %s failed: %s", env_MASTER_ID, err.Error())
	}
	MasterID = uint64(adminId)

	proxyURLStr := os.Getenv(env_PROXY_URL)
	if proxyURLStr != "" {
		if url, err := url.Parse(proxyURLStr); err == nil {
			ProxyURL = *url
		} else {
			log.Fatalf("Parse %s failed: %s", env_PROXY_URL, proxyURLStr)
		}
	}
}
