package qbot

import (
	"encoding/json"
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
	config        *Config
	conn          *websocket.Conn
	retryCount    int
	stopChan      chan bool
	pendingEcho   sync.Map
	mutex         sync.Mutex
	eventHandlers struct {
		onMessage func(c *Client, msg *Message)
	}
}

type MsgType int

const (
	Text    MsgType = 0
	At      MsgType = 1
	Face    MsgType = 2
	Image   MsgType = 3
	Record  MsgType = 4
	Reply   MsgType = 5
	File    MsgType = 6
	Forward MsgType = 7
	Json    MsgType = 8

	Other MsgType = -1
)

type MsgItem struct {
	Type    MsgType
	Content string
}

type Message struct {
	GroupID  uint64
	UserID   uint64
	Nickname string
	Card     string
	Role     string
	Time     uint64
	MsgID    uint64
	Raw      string
	Content  string
	Array    []MsgItem
}

type messageJson struct {
	GroupID   uint64 `json:"group_id"`
	Time      uint64 `json:"time"`
	MessageID uint64 `json:"message_id"`
	Sender    struct {
		UserID   uint64 `json:"user_id"`
		Nickname string `json:"nickname"`
		Card     string `json:"card"`
		Role     string `json:"role"`
	} `json:"sender"`
	RawMessage string `json:"raw_message"`
	Message    []struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	} `json:"message"`
}

type pendingResponse struct {
	ch    chan *cqResponse
	timer *time.Timer
}

type cqRequest struct {
	Action string         `json:"action"`
	Params map[string]any `json:"params"`
	Echo   string         `json:"echo,omitempty"`
}

type cqResponse struct {
	Status  string `json:"status"`
	Retcode int    `json:"retcode"`
	Data    struct {
		MessageId uint64 `json:"message_id"`
	}
	Message string `json:"message"`
	Wording string `json:"wording"`
	Echo    string `json:"echo"`
}
