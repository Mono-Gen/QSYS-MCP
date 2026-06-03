package qsys

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"qsys-mcp/config"
)

const (
	DefaultQrcPort        = 1710
	DefaultTimeout        = 10 * time.Second
	MaxReconnectDelay     = 30 * time.Second
	BaseReconnectDelay    = 500 * time.Millisecond
	KeepAliveInterval     = 55 * time.Second
)

type QrcClient struct {
	host       string
	port       int
	user       string
	pin        string
	timeout    time.Duration
	socket     net.Conn
	connMutex  sync.RWMutex
	connected  bool
	reconnecting bool
	destroyed  bool
	
	nextID     int
	idMutex    sync.Mutex
	pending    map[int]chan *JsonRpcResponse
	pendMutex  sync.Mutex
	
	stopKeepAlive chan struct{}
}

func NewQrcClient(host string, port int) *QrcClient {
	if port == 0 {
		port = DefaultQrcPort
	}
	return &QrcClient{
		host:    host,
		port:    port,
		timeout: DefaultTimeout,
		pending: make(map[int]chan *JsonRpcResponse),
	}
}

func (c *QrcClient) SetAuth(user, pin string) {
	c.user = user
	c.pin = pin
}

func (c *QrcClient) IsConnected() bool {
	c.connMutex.RLock()
	defer c.connMutex.RUnlock()
	return c.connected
}

func (c *QrcClient) Connect() error {
	c.connMutex.Lock()
	if c.connected {
		c.connMutex.Unlock()
		return nil
	}
	if c.reconnecting {
		c.connMutex.Unlock()
		// Wait a short time until connection completes or times out
		for i := 0; i < 100; i++ {
			time.Sleep(50 * time.Millisecond)
			c.connMutex.RLock()
			isConnected := c.connected
			c.connMutex.RUnlock()
			if isConnected {
				return nil
			}
		}
		return errors.New("concurrent connection attempt in progress")
	}
	c.connMutex.Unlock()
	return c.performConnect()
}

func (c *QrcClient) performConnect() error {
	address := fmt.Sprintf("%s:%d", c.host, c.port)
	config.DebugLog("Connecting to QRC at %s...", address)
	
	dialer := net.Dialer{Timeout: c.timeout}
	conn, err := dialer.Dial("tcp", address)
	if err != nil {
		return err
	}

	c.connMutex.Lock()
	c.socket = conn
	c.connected = true
	c.destroyed = false
	c.stopKeepAlive = make(chan struct{})
	c.connMutex.Unlock()

	// Start reading loop (needed beforehand to receive Call responses)
	go c.readLoop(conn)

	// Authentication via Logon
	if c.user != "" || c.pin != "" {
		config.DebugLog("Authenticating QRC connection...")
		_, err := c.Call("Logon", map[string]interface{}{
			"User": c.user,
			"Pin":  c.pin,
		})
		if err != nil {
			_ = c.Disconnect()
			return fmt.Errorf("QRC logon failed: %w", err)
		}
		config.DebugLog("QRC connection authenticated successfully")
	}

	// Start keep-alive loop
	go c.keepAliveLoop()

	config.DebugLog("Connected to QRC at %s", address)
	return nil
}

func (c *QrcClient) readLoop(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		// QRC is null-character (\x00) delimited
		data, err := reader.ReadBytes(0)
		if err != nil {
			c.handleDisconnect(err)
			return
		}

		// Remove trailing null-character
		if len(data) > 0 && data[len(data)-1] == 0 {
			data = data[:len(data)-1]
		}

		trimmed := bytes.TrimSpace(data)
		if len(trimmed) == 0 {
			continue
		}

		var res JsonRpcResponse
		if err := json.Unmarshal(trimmed, &res); err != nil {
			// Ignore parse errors and continue
			continue
		}

		if res.ID == nil {
			// Ignore notifications, etc.
			continue
		}

		var idVal int
		switch idTyped := res.ID.(type) {
		case float64:
			idVal = int(idTyped)
		case int:
			idVal = idTyped
		default:
			// Ignore unknown ID types
			continue
		}

		c.pendMutex.Lock()
		ch, found := c.pending[idVal]
		c.pendMutex.Unlock()

		if found {
			ch <- &res
		}
	}
}

func (c *QrcClient) handleDisconnect(err error) {
	c.connMutex.Lock()
	if !c.connected {
		c.connMutex.Unlock()
		return
	}
	c.connected = false
	if c.socket != nil {
		c.socket.Close()
	}
	close(c.stopKeepAlive)
	c.connMutex.Unlock()

	config.DebugLog("QRC disconnected: %v", err)

	// Mark all pending requests as failed
	c.pendMutex.Lock()
	for id, ch := range c.pending {
		ch <- &JsonRpcResponse{
			Jsonrpc: "2.0",
			ID:      id,
			Error: &JsonRpcError{
				Code:    -32603,
				Message: fmt.Sprintf("connection lost: %v", err),
			},
		}
	}
	c.pending = make(map[int]chan *JsonRpcResponse)
	c.pendMutex.Unlock()

	c.connMutex.RLock()
	destroyed := c.destroyed
	c.connMutex.RUnlock()

	if !destroyed {
		go c.scheduleReconnect()
	}
}

func (c *QrcClient) scheduleReconnect() {
	c.connMutex.Lock()
	if c.reconnecting || c.destroyed {
		c.connMutex.Unlock()
		return
	}
	c.reconnecting = true
	c.connMutex.Unlock()

	delay := BaseReconnectDelay
	for {
		c.connMutex.RLock()
		destroyed := c.destroyed
		c.connMutex.RUnlock()
		if destroyed {
			break
		}

		time.Sleep(delay)
		
		err := c.performConnect()
		if err == nil {
			c.connMutex.Lock()
			c.reconnecting = false
			c.connMutex.Unlock()
			return
		}

		delay *= 2
		if delay > MaxReconnectDelay {
			delay = MaxReconnectDelay
		}
	}

	c.connMutex.Lock()
	c.reconnecting = false
	c.connMutex.Unlock()
}

func (c *QrcClient) keepAliveLoop() {
	ticker := time.NewTicker(KeepAliveInterval)
	defer ticker.Stop()

	c.connMutex.RLock()
	stopCh := c.stopKeepAlive
	c.connMutex.RUnlock()

	for {
		select {
		case <-ticker.C:
			if c.IsConnected() {
				_, _ = c.Call("NoOp", nil)
			}
		case <-stopCh:
			return
		}
	}
}

func (c *QrcClient) Call(method string, params interface{}) (interface{}, error) {
	if !c.IsConnected() {
		if err := c.Connect(); err != nil {
			return nil, fmt.Errorf("qrc not connected: %w", err)
		}
	}

	c.idMutex.Lock()
	c.nextID++
	id := c.nextID
	c.idMutex.Unlock()

	req := JsonRpcRequest{
		Jsonrpc: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	// QRC is null-character (\x00) delimited
	payload = append(payload, 0)

	ch := make(chan *JsonRpcResponse, 1)
	c.pendMutex.Lock()
	c.pending[id] = ch
	c.pendMutex.Unlock()

	c.connMutex.RLock()
	socket := c.socket
	c.connMutex.RUnlock()

	if socket == nil {
		c.pendMutex.Lock()
		delete(c.pending, id)
		c.pendMutex.Unlock()
		return nil, errors.New("socket is nil")
	}

	_, err = socket.Write(payload)
	if err != nil {
		c.pendMutex.Lock()
		delete(c.pending, id)
		c.pendMutex.Unlock()
		return nil, err
	}

	select {
	case res := <-ch:
		if res.Error != nil {
			return nil, fmt.Errorf("qrc error %d: %s", res.Error.Code, res.Error.Message)
		}
		return res.Result, nil
	case <-time.After(c.timeout):
		c.pendMutex.Lock()
		delete(c.pending, id)
		c.pendMutex.Unlock()
		return nil, fmt.Errorf("request timed out: %s (id=%d)", method, id)
	}
}

func (c *QrcClient) Disconnect() error {
	c.connMutex.Lock()
	c.destroyed = true
	c.connected = false
	if c.socket != nil {
		c.socket.Close()
		c.socket = nil
	}
	c.connMutex.Unlock()
	return nil
}
