package qbot

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Config struct {
	Address      string        `json:"address"`
	Reconnect    time.Duration `json:"reconnect"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
}

type Client struct {
	config      *Config
	conn        *websocket.Conn
	retryCount  int
	stopChan    chan bool
	pendingEcho sync.Map
	mutex       sync.Mutex
}

type pendingResponse struct {
	ch    chan *CQResponse
	timer *time.Timer
}

type CQRequest struct {
	Action string         `json:"action"`
	Params map[string]any `json:"params"`
	Echo   string         `json:"echo,omitempty"`
}

type CQResponse struct {
	Status  string `json:"status"`
	Retcode int    `json:"retcode"`
	Data    struct {
		MessageId uint64 `json:"message_id"`
	}
	Message string `json:"message"`
	Wording string `json:"wording"`
	Echo    string `json:"echo"`
}

type PushMessage struct {
	SelfID      int64  `json:"self_id"`
	UserID      int64  `json:"user_id"`
	Time        int64  `json:"time"`
	MessageID   int64  `json:"message_id"`
	MessageType string `json:"message_type"`
	Sender      struct {
		UserID   int64  `json:"user_id"`
		Nickname string `json:"nickname"`
		Card     string `json:"card"`
	} `json:"sender"`
	RawMessage string `json:"raw_message"`
	Message    []struct {
		Type string `json:"type"`
		Data struct {
			Text string `json:"text"`
		} `json:"data"`
	} `json:"message"`
}

type HeartbeatMessage struct {
	Time     int64 `json:"time"`
	SelfID   int64 `json:"self_id"`
	Interval int64 `json:"interval"`
	Status   struct {
		Online bool `json:"online"`
		Good   bool `json:"good"`
	} `json:"status"`
}
