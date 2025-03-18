// qbot/qbot.go
package qbot

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"go-hurobot/config"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func init() {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.PsqlHost, strconv.Itoa(int(config.PsqlPort)), config.PsqlUser, config.PsqlPassword, config.PsqlDbName)
	if err := initPsqlDB(dsn); err != nil {
		log.Fatalln(err)
	}
}

func NewClient() *Client {
	client := &Client{
		config: &Config{
			Address:      config.NapcatWSURL,
			Reconnect:    3 * time.Second,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		retryCount: 0,
		stopChan:   make(chan bool),
	}
	go client.connect()
	return client
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
	if c.stopChan != nil {
		close(c.stopChan)
	}
}

func (c *Client) connect() {
	for {
		select {
		case <-c.stopChan:
			return
		default:
			// TODO
		}

		dialer := websocket.Dialer{
			ReadBufferSize:   4096,
			WriteBufferSize:  4096,
			HandshakeTimeout: c.config.ReadTimeout,
		}
		conn, _, err := dialer.Dial(c.config.Address, nil)
		if err != nil {
			log.Printf("Connect failed (%d): %v", c.retryCount+1, err)
			c.retryCount++
			time.Sleep(c.config.Reconnect)
			continue
		}

		conn.SetPongHandler(func(string) error {
			c.retryCount = 0
			return nil
		})

		c.conn = conn
		log.Println("Connected to NapCat")

		go c.messageHandler()
		return
	}
}

func (c *Client) messageHandler() {
	defer func() {
		if c.conn != nil {
			c.conn.Close()
		}
	}()

	for {
		// Receive message
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("read error: %v", err)
			c.reconnect()
			return
		}

		// Unmarshal to map
		jsonMap := make(map[string]any)
		if err := json.Unmarshal(msg, &jsonMap); err != nil {
			log.Printf("parse message error: %v", err)
			continue
		}

		if jsonMap["echo"] != nil {
			// Response to sent message
			var resp cqResponse
			if err := json.Unmarshal(msg, &resp); err == nil {
				c.mutex.Lock()
				if val, ok := c.pendingEcho.Load(resp.Echo); ok {
					pr := val.(*pendingResponse)
					pr.timer.Stop()
					pr.ch <- &resp
					c.pendingEcho.Delete(resp.Echo)
				}
				c.mutex.Unlock()
			}
		} else if postType, exists := jsonMap["post_type"]; exists {
			// Server-initiated push
			if str, ok := postType.(string); ok && str != "" {
				c.handleEvents(&str, &msg, &jsonMap)
			}
		}
	}
}

func (c *Client) reconnect() {
	if c.stopChan != nil {
		close(c.stopChan)
	}
	c.stopChan = make(chan bool)
	c.connect()
}

func (c *Client) sendJson(req *cqRequest) error {
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return err
	}
	if c.conn == nil {
		return fmt.Errorf("connection not ready")
	}
	if err := c.conn.WriteMessage(websocket.TextMessage, jsonBytes); err != nil {
		return err
	}
	return nil
}

func (c *Client) sendJsonWithEcho(req *cqRequest) (*cqResponse, error) {
	// Generate echo key
	echo := uuid.New().String()
	req.Echo = echo

	respCh := make(chan *cqResponse, 1)
	timeout := time.NewTimer(5 * time.Second)
	defer timeout.Stop()

	// Save the key to pendingEcho
	c.mutex.Lock()
	c.pendingEcho.Store(echo, &pendingResponse{
		ch:    respCh,
		timer: timeout,
	})
	c.mutex.Unlock()

	// Send request
	if err := c.sendJson(req); err != nil {
		return nil, err
	}

	// Wait for response
	select {
	case resp := <-respCh:
		if resp == nil {
			return nil, fmt.Errorf("response channel closed")
		} else {
			log.Printf("Sent message: %v", req.Params)
		}
		return resp, nil
	case <-timeout.C:
		c.mutex.Lock()
		c.pendingEcho.Delete(echo)
		c.mutex.Unlock()
		return nil, fmt.Errorf("wait response timeout")
	}
}
