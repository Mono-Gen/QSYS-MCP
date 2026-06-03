package qsys

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"qsys-mcp/config"
)

type EcpClient struct {
	host              string
	port              int
	socket            net.Conn
	connMutex         sync.RWMutex
	connected         bool
	reconnectAttempts int
	
	lineBuffer        []string
	bufferMutex       sync.Mutex
	lineResolve       chan string
	
	maxReconnectDelay time.Duration
	baseReconnectDelay time.Duration
}

func NewEcpClient(host string, port int) *EcpClient {
	if port == 0 {
		port = 1702
	}
	return &EcpClient{
		host:              host,
		port:              port,
		maxReconnectDelay: 5 * time.Second,
		baseReconnectDelay: 100 * time.Millisecond,
		lineResolve:       make(chan string, 1),
	}
}

func (c *EcpClient) IsConnected() bool {
	c.connMutex.RLock()
	defer c.connMutex.RUnlock()
	return c.connected
}

func (c *EcpClient) Connect() error {
	c.connMutex.RLock()
	if c.connected && c.socket != nil {
		c.connMutex.RUnlock()
		return nil
	}
	c.connMutex.RUnlock()

	return c.connectWithRetry()
}

func (c *EcpClient) connectWithRetry() error {
	var lastErr error
	for !c.IsConnected() {
		err := c.performConnect()
		if err == nil {
			c.reconnectAttempts = 0
			return nil
		}
		lastErr = err

		delay := time.Duration(float64(c.baseReconnectDelay) * math.Pow(2, float64(c.reconnectAttempts)))
		if delay > c.maxReconnectDelay {
			delay = c.maxReconnectDelay
		}
		c.reconnectAttempts++

		if c.reconnectAttempts > 2 {
			return lastErr
		}

		time.Sleep(delay)
	}
	return nil
}

func (c *EcpClient) performConnect() error {
	address := fmt.Sprintf("%s:%d", c.host, c.port)
	config.DebugLog("Connecting to ECP at %s...", address)

	dialer := net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.Dial("tcp", address)
	if err != nil {
		return err
	}

	c.connMutex.Lock()
	c.socket = conn
	c.connected = true
	c.connMutex.Unlock()

	go c.readLoop(conn)

	config.DebugLog("Connected to ECP at %s", address)
	return nil
}

func (c *EcpClient) readLoop(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 {
			select {
			case c.lineResolve <- line:
				// Successfully sent to the pending channel
			default:
				// Add to buffer if no channel is waiting
				c.bufferMutex.Lock()
				c.lineBuffer = append(c.lineBuffer, line)
				c.bufferMutex.Unlock()
			}
		}
	}

	c.connMutex.Lock()
	c.connected = false
	if c.socket != nil {
		c.socket.Close()
		c.socket = nil
	}
	c.connMutex.Unlock()
}

func (c *EcpClient) Disconnect() error {
	c.connMutex.Lock()
	c.connected = false
	if c.socket != nil {
		c.socket.Close()
		c.socket = nil
	}
	c.connMutex.Unlock()

	c.bufferMutex.Lock()
	c.lineBuffer = []string{}
	c.bufferMutex.Unlock()
	return nil
}

func (c *EcpClient) sendCommand(command string) (string, error) {
	if !c.IsConnected() {
		if err := c.Connect(); err != nil {
			return "", err
		}
	}

	c.connMutex.RLock()
	socket := c.socket
	c.connMutex.RUnlock()

	if socket == nil {
		return "", errors.New("ecp socket is nil")
	}

	// Send command
	_, err := socket.Write([]byte(command + "\r\n"))
	if err != nil {
		return "", err
	}

	// Check if there is data in the buffer
	c.bufferMutex.Lock()
	if len(c.lineBuffer) > 0 {
		line := c.lineBuffer[0]
		c.lineBuffer = c.lineBuffer[1:]
		c.bufferMutex.Unlock()
		return line, nil
	}
	c.bufferMutex.Unlock()

	// Wait for response (5s timeout)
	select {
	case line := <-c.lineResolve:
		return line, nil
	case <-time.After(5 * time.Second):
		return "", errors.New("ecp command timeout")
	}
}

func (c *EcpClient) GetControl(name string) (*EcpControlResult, error) {
	response, err := c.sendCommand(fmt.Sprintf("cgc %s", name))
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(response, "ERR") {
		return nil, fmt.Errorf("ecp error: %s", response)
	}

	parts := strings.Split(response, ",")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid ecp response: %s", response)
	}

	val, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse value: %w", err)
	}

	pos, err := strconv.ParseFloat(parts[len(parts)-1], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse position: %w", err)
	}

	strVal := strings.Join(parts[1:len(parts)-1], ",")

	return &EcpControlResult{
		Value:    val,
		String:   strVal,
		Position: pos,
	}, nil
}

func (c *EcpClient) SetControlValue(name string, value float64) error {
	response, err := c.sendCommand(fmt.Sprintf("csv %s %g", name, value))
	if err != nil {
		return err
	}
	if strings.HasPrefix(response, "ERR") {
		return fmt.Errorf("ecp error: %s", response)
	}
	return nil
}

func (c *EcpClient) SetControlPosition(name string, position float64) error {
	if position < 0.0 || position > 1.0 {
		return fmt.Errorf("position must be between 0.0 and 1.0, got %f", position)
	}
	response, err := c.sendCommand(fmt.Sprintf("csp %s %g", name, position))
	if err != nil {
		return err
	}
	if strings.HasPrefix(response, "ERR") {
		return fmt.Errorf("ecp error: %s", response)
	}
	return nil
}

func (c *EcpClient) GetControlString(name string) (string, error) {
	response, err := c.sendCommand(fmt.Sprintf("cgs %s", name))
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(response, "ERR") {
		return "", fmt.Errorf("ecp error: %s", response)
	}
	return response, nil
}

func (c *EcpClient) LoadSnapshotByName(snapshotName string) error {
	response, err := c.sendCommand(fmt.Sprintf("ssl %s", snapshotName))
	if err != nil {
		return err
	}
	if strings.HasPrefix(response, "ERR") {
		return fmt.Errorf("ecp error: %s", response)
	}
	return nil
}

func (c *EcpClient) LoadSnapshotByPosition(bank int, number int) error {
	response, err := c.sendCommand(fmt.Sprintf("ssa %d %d", bank, number))
	if err != nil {
		return err
	}
	if strings.HasPrefix(response, "ERR") {
		return fmt.Errorf("ecp error: %s", response)
	}
	return nil
}
