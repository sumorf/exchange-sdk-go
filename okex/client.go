package okex

import (
	"bytes"
	"compress/flate"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/blockcdn-go/exchange-sdk-go/config"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/gotoxu/log/core"
	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

// WSSClient 提供okex API调用的客户端
type WSSClient struct {
	config config.Config
	conns  map[string]*websocket.Conn
	logger core.Logger

	closed  bool
	closeMu sync.Mutex
	once    sync.Once

	shouldQuit chan struct{}
	retry      chan string
	done       chan struct{}

	events []Event
}

// NewWSSClient 创建一个新的Websocket client
func NewWSSClient(config *config.Config) *WSSClient {
	cfg := defaultConfig()
	if config != nil {
		cfg.MergeIn(config)
	}

	return &WSSClient{
		config:     *cfg,
		conns:      make(map[string]*websocket.Conn),
		shouldQuit: make(chan struct{}),
		retry:      make(chan string),
		done:       make(chan struct{}),
	}
}

// SetLogger 设置日志器
func (c *WSSClient) SetLogger(logger core.Logger) {
	c.logger = logger
}

// Subscribe 订阅
func (c *WSSClient) Subscribe(events []Event) {
	c.events = events
}

// QuerySpot 负责订阅现货行情数据
func (c *WSSClient) QuerySpot() (<-chan []byte, error) {
	cid, conn, err := c.connect("/websocket")
	if err != nil {
		return nil, err
	}

	err = c.subscribeSpot(conn)
	if err != nil {
		c.Close()
		return nil, err
	}

	result := make(chan []byte)
	go c.start(cid, "/websocket", result)

	return result, nil
}

func (c *WSSClient) log(level core.Level, v ...interface{}) {
	if c.logger != nil {
		c.logger.Log(level, v...)
	}
}

func (c *WSSClient) logf(level core.Level, format string, v ...interface{}) {
	if c.logger != nil {
		c.logger.Logf(level, format, v...)
	}
}

func (c *WSSClient) logln(level core.Level, v ...interface{}) {
	if c.logger != nil {
		c.logger.Logln(level, v...)
	}
}

func (c *WSSClient) subscribeSpot(conn *websocket.Conn) error {
	for _, v := range c.events {
		err := conn.WriteJSON(v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *WSSClient) start(cid, path string, msgCh chan<- []byte) {
	go c.query(cid, msgCh)

	ticker := time.NewTicker(*c.config.PingDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			conn := c.conns[cid]
			conn.WriteMessage(websocket.TextMessage, []byte("{'event':'ping'}"))
		case cid := <-c.retry:
			delete(c.conns, cid)
			c.reconnect(path, msgCh)
		case <-c.shouldQuit:
			c.shutdown()
			return
		}
	}
}

func (c *WSSClient) shutdown() {
	c.closeMu.Lock()
	c.closed = true
	c.closeMu.Unlock()

	for _, conn := range c.conns {
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		conn.Close()
	}

	close(c.done)
}

func (c *WSSClient) reconnect(path string, msgCh chan<- []byte) {
	cid, conn, err := c.connect(path)
	if err != nil {
		return
	}

	err = c.subscribeSpot(conn)
	if err != nil {
		c.closeMu.Lock()
		defer c.closeMu.Unlock()

		if !c.closed {
			c.retry <- cid
		}
		return
	}

	go c.query(cid, msgCh)
}

func (c *WSSClient) query(cid string, msgCh chan<- []byte) {
	conn, ok := c.conns[cid]
	if !ok {
		return
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			c.closeMu.Lock()
			defer c.closeMu.Unlock()

			if !c.closed {
				c.retry <- cid
			}
			return
		}

		if strings.Contains(string(msg), "pong") {
			continue
		}

		buf := bytes.NewBuffer(msg)
		z := flate.NewReader(buf)
		m, _ := ioutil.ReadAll(z)

		// 作为sdk开发者来说，我们并不知晓调用方需要哪些数据，因此这里不做过滤，而在调用方再做过滤
		// 从代码致性效率来说，这里反序列化一次，然后将消息传给调用方，调用方仍然需要反序列化，因此直接在调用方做反序列化并过滤
		// 而sdk中不做处理
		// var subrsp [1]struct {
		// 	Data struct {
		// 		Result string `json:"result"`
		// 	} `json:"data"`
		// }
		// if e := json.Unmarshal(m, &subrsp); e != nil {
		// 	// 订阅请求的回复，不包含数据，直接忽略
		// 	log.Print("ignore subscribe respone.")
		// 	continue
		// }
		msgCh <- m
	}
}

// Close 向服务端发起关闭操作
func (c *WSSClient) Close() {
	if c.conns == nil {
		return
	}

	close(c.shouldQuit)

	select {
	case <-c.done:
	case <-time.After(time.Second):
	}
}

func (c *WSSClient) connect(path string) (string, *websocket.Conn, error) {
	u := url.URL{Scheme: "wss", Host: *c.config.WSSHost, Path: path}
	conn, _, err := c.config.WSSDialer.Dial(u.String(), nil)
	if err == nil {
		u := uuid.New().String()
		c.conns[u] = conn
		return u, conn, nil
	}

	if err == websocket.ErrBadHandshake {
		return "", nil, err
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			conn, _, err := c.config.WSSDialer.Dial(u.String(), nil)
			if err == nil {
				u := uuid.New().String()
				c.conns[u] = conn
				return u, conn, nil
			}
		case <-c.shouldQuit:
			return "", nil, errors.New("Connection is closing")
		}
	}
}

// Client ...
type Client struct {
	config config.Config
}

// NewClient 创建一个新的client
func NewClient(config *config.Config) *Client {
	cfg := defaultConfig()
	if config != nil {
		cfg.MergeIn(config)
	}

	return &Client{config: *cfg}
}

func (c *Client) doHTTP(method, path string,
	reqarg interface{}, out interface{}, extraheader map[string]string) error {
	url := "http://"
	if *c.config.UseSSL {
		url = "https://"
	}
	url += *c.config.RESTHost + path

	arg := ""
	if method == "POST" {
		bytesParams, _ := json.Marshal(reqarg)
		arg = string(bytesParams)
	}
	req, e := http.NewRequest(method, url, strings.NewReader(arg))
	if e != nil {
		return e
	}
	if method == "GET" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept-Language", "zh-cn")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) "+
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.71 Safari/537.36")
	for k, v := range extraheader {
		req.Header.Set(k, v)
	}
	resp, re := c.config.HTTPClient.Do(req)
	if e != nil {
		return re
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求失败，响应码：%d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("okex http message:%s\n", string(body))
	extra.RegisterFuzzyDecoders()

	err = jsoniter.Unmarshal(body, out)
	if err != nil {
		return err
	}
	return nil
}
