package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

type ToolHandler func(params map[string]interface{}) (string, error)

type Server struct {
	tools    []Tool
	handlers map[string]ToolHandler
	mutex    sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		handlers: make(map[string]ToolHandler),
	}
}

func (s *Server) RegisterTool(tool Tool, handler ToolHandler) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.tools = append(s.tools, tool)
	s.handlers[tool.Name] = handler
}

type Request struct {
	Jsonrpc string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id,omitempty"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

type Response struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type InitializeResult struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ServerInfo      ServerInfo             `json:"serverInfo"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ToolListResult struct {
	Tools []Tool `json:"tools"`
}

type ToolCallResult struct {
	Content []TextContent `json:"content"`
	IsError bool          `json:"isError"`
}

type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (s *Server) Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			s.sendError(out, nil, -32700, "Parse error: "+err.Error())
			continue
		}

		s.handleRequest(out, &req)
	}
}

func (s *Server) handleRequest(out io.Writer, req *Request) {
	switch req.Method {
	case "initialize":
		res := Response{
			Jsonrpc: "2.0",
			ID:      req.ID,
			Result: InitializeResult{
				ProtocolVersion: "2024-11-05",
				Capabilities: map[string]interface{}{
					"tools": map[string]interface{}{},
				},
				ServerInfo: ServerInfo{
					Name:    "q-sys-mcp",
					Version: "0.1.0",
				},
			},
		}
		s.sendResponse(out, &res)

	case "notifications/initialized":
		// No response required

	case "tools/list":
		s.mutex.RLock()
		tools := s.tools
		s.mutex.RUnlock()

		res := Response{
			Jsonrpc: "2.0",
			ID:      req.ID,
			Result: ToolListResult{
				Tools: tools,
			},
		}
		s.sendResponse(out, &res)

	case "tools/call":
		toolName, _ := req.Params["name"].(string)
		args, _ := req.Params["arguments"].(map[string]interface{})

		s.mutex.RLock()
		handler, ok := s.handlers[toolName]
		s.mutex.RUnlock()

		if !ok {
			s.sendError(out, req.ID, -32601, fmt.Sprintf("Method not found: %s", toolName))
			return
		}

		go func() {
			text, err := handler(args)
			var res Response
			if err != nil {
				res = Response{
					Jsonrpc: "2.0",
					ID:      req.ID,
					Result: ToolCallResult{
						Content: []TextContent{
							{Type: "text", Text: "Error: " + err.Error()},
						},
						IsError: true,
					},
				}
			} else {
				res = Response{
					Jsonrpc: "2.0",
					ID:      req.ID,
					Result: ToolCallResult{
						Content: []TextContent{
							{Type: "text", Text: text},
						},
						IsError: false,
					},
				}
			}
			s.sendResponse(out, &res)
		}()

	default:
		if req.ID != nil {
			s.sendError(out, req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method))
		}
	}
}

func (s *Server) sendResponse(out io.Writer, res *Response) {
	data, err := json.Marshal(res)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling response: %v\n", err)
		return
	}
	out.Write(data)
	out.Write([]byte("\n"))
}

func (s *Server) sendError(out io.Writer, id interface{}, code int, msg string) {
	res := Response{
		Jsonrpc: "2.0",
		ID:      id,
		Error: &Error{
			Code:    code,
			Message: msg,
		},
	}
	s.sendResponse(out, &res)
}
