package qsys

type ConnectionState string

const (
	Disconnected ConnectionState = "disconnected"
	Connecting   ConnectionState = "connecting"
	Connected    ConnectionState = "connected"
	Error        ConnectionState = "error"
)

type CoreAlias string

type CoreConfig struct {
	Alias   CoreAlias
	Host    string
	QrcPort int // default 1710
	EcpPort int // default 1702
	User    string
	Pin     string
}

type IQrcClient interface {
	IsConnected() bool
	Connect() error
	Disconnect() error
	Call(method string, params interface{}) (interface{}, error)
}

type JsonRpcRequest struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type JsonRpcResponse struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *JsonRpcError   `json:"error,omitempty"`
}

type JsonRpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type EcpControlResult struct {
	Value    float64 `json:"value"`
	String   string  `json:"string"`
	Position float64 `json:"position"`
}

type CoreStatus struct {
	Name       string `json:"name"`
	DesignName string `json:"designName"`
	IsRunning  bool   `json:"isRunning"`
	Platform   string `json:"platform,omitempty"`
	Uptime     int    `json:"uptime,omitempty"`
}
