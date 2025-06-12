package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"

	"github.com/miyamo2/qilin"
	"github.com/miyamo2/qilin/transport"
	"github.com/stretchr/testify/suite"
)

type StreamableTestSuite struct {
	suite.Suite
	mu      sync.Mutex
	cancel  context.CancelFunc
	address string
}

func (s *StreamableTestSuite) BeforeTest(_, _ string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	q := NewQilin(s.T())
	listener, err := net.Listen("tcp", "localhost:0")
	s.Require().NoError(err, "failed to create tcp listener")

	s.address = listener.Addr().String()

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	ready := make(chan struct{}, 1)
	go func() {
		streamable := transport.NewStreamable(transport.StreamableWithNetListener(listener))
		q.Start(
			qilin.StartWithReadySignal(ready),
			qilin.StartWithContext(ctx),
			qilin.StartWithListener(streamable))
	}()
	<-ready
}

func (s *StreamableTestSuite) AfterTest(_, _ string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cancel()
	s.cancel = nil
}

func TestStreamableTestSuite(t *testing.T) {
	suite.Run(t, new(StreamableTestSuite))
}

// TestStreamableTestSuite_Initialize_Success tests successful initialization process via HTTP
func (s *StreamableTestSuite) TestStreamableTestSuite_Initialize_Success() {
	params := map[string]any{
		"protocolVersion": qilin.LatestProtocolVersion,
		"capabilities": map[string]any{
			"experimental": map[string]any{},
		},
		"clientInfo": map[string]any{
			"name":    "test-client",
			"version": "1.0.0",
		},
	}

	req := NewJSONRPCRequest(s.T(), qilin.MethodInitialize, params)
	reqBytes, err := json.Marshal(req)
	s.Require().NoError(err)

	url := fmt.Sprintf("http://%s/mcp", s.address)
	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBytes))
	s.Require().NoError(err)
	defer resp.Body.Close()

	sessionID := SessionIDFromResponse(s.T(), resp)
	s.Require().NotEmpty(sessionID)

	var responseData []byte
	for event := range StreamIterFromResponse(s.T(), resp) {
		responseData = event
		break
	}

	response := JSONRPCResponseFromBytes(s.T(), responseData)

	s.Require().Equal(req.ID, response.ID)
	s.Require().Nil(response.Error)
	s.Require().NotNil(response.Result)

	var result map[string]any
	err = json.Unmarshal(response.Result, &result)
	s.Require().NoError(err)

	protocolVersion, ok := result["protocolVersion"].(string)
	s.Require().True(ok)
	s.Require().Equal(qilin.LatestProtocolVersion, protocolVersion)

	capabilities, ok := result["capabilities"].(map[string]any)
	s.Require().True(ok)
	s.Require().NotNil(capabilities)

	serverInfo, ok := result["serverInfo"].(map[string]any)
	s.Require().True(ok)
	s.Require().NotNil(serverInfo)

	name, ok := serverInfo["name"].(string)
	s.Require().True(ok)
	s.Require().Equal("beer_hall", name)

	version, ok := serverInfo["version"].(string)
	s.Require().True(ok)
	s.Require().Equal("1.0.0", version)
}

// TestStreamableTestSuite_Initialize_ProtocolVersionFallback tests protocol version fallback
func (s *StreamableTestSuite) TestStreamableTestSuite_Initialize_ProtocolVersionFallback() {
	params := map[string]any{
		"protocolVersion": "unsupported-version-1.0.0",
		"capabilities": map[string]any{
			"experimental": map[string]any{},
		},
		"clientInfo": map[string]any{
			"name":    "test-client",
			"version": "1.0.0",
		},
	}
	req := NewJSONRPCRequest(s.T(), qilin.MethodInitialize, params)
	reqBytes, err := json.Marshal(req)
	s.Require().NoError(err)

	url := fmt.Sprintf("http://%s/mcp", s.address)
	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBytes))
	s.Require().NoError(err)
	defer resp.Body.Close()

	sessionID := SessionIDFromResponse(s.T(), resp)
	s.Require().NotEmpty(sessionID)

	var responseData []byte
	for event := range StreamIterFromResponse(s.T(), resp) {
		responseData = event
		break
	}

	response := JSONRPCResponseFromBytes(s.T(), responseData)

	s.Require().Equal(req.ID, response.ID)
	s.Require().Nil(response.Error)
	s.Require().NotNil(response.Result)

	var result map[string]any
	err = json.Unmarshal(response.Result, &result)
	s.Require().NoError(err)

	protocolVersion, ok := result["protocolVersion"].(string)
	s.Require().True(ok)
	s.Require().Equal(qilin.LatestProtocolVersion, protocolVersion)

	capabilities, ok := result["capabilities"].(map[string]any)
	s.Require().True(ok)
	s.Require().NotNil(capabilities)

	serverInfo, ok := result["serverInfo"].(map[string]any)
	s.Require().True(ok)
	s.Require().NotNil(serverInfo)

	name, ok := serverInfo["name"].(string)
	s.Require().True(ok)
	s.Require().Equal("beer_hall", name)

	version, ok := serverInfo["version"].(string)
	s.Require().True(ok)
	s.Require().Equal("1.0.0", version)
}

// TestStreamableTestSuite_Ping_Success tests successful ping request after session establishment
func (s *StreamableTestSuite) TestStreamableTestSuite_Ping_Success() {
	params := map[string]any{
		"protocolVersion": qilin.LatestProtocolVersion,
		"capabilities": map[string]any{
			"experimental": map[string]any{},
		},
		"clientInfo": map[string]any{
			"name":    "test-client",
			"version": "1.0.0",
		},
	}
	initReq := NewJSONRPCRequest(s.T(), qilin.MethodInitialize, params)
	initReqBytes, err := json.Marshal(initReq)
	s.Require().NoError(err)

	url := fmt.Sprintf("http://%s/mcp", s.address)
	initResp, err := http.Post(url, "application/json", bytes.NewReader(initReqBytes))
	s.Require().NoError(err)
	defer initResp.Body.Close()

	sessionID := SessionIDFromResponse(s.T(), initResp)
	s.Require().NotEmpty(sessionID)

	pingReq := NewJSONRPCRequest(s.T(), qilin.MethodPing, nil)
	pingReqBytes, err := json.Marshal(pingReq)
	s.Require().NoError(err)

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(pingReqBytes))
	s.Require().NoError(err)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set(transport.MCPSessionID, sessionID)

	client := &http.Client{}
	pingResp, err := client.Do(httpReq)
	s.Require().NoError(err)
	defer pingResp.Body.Close()

	s.Require().Equal(http.StatusOK, pingResp.StatusCode)

	responseData, err := io.ReadAll(pingResp.Body)
	s.Require().NoError(err)

	pingResponse := JSONRPCResponseFromBytes(s.T(), responseData)

	s.Require().Equal(pingReq.ID, pingResponse.ID)
	s.Require().Nil(pingResponse.Error)
	s.Require().NotNil(pingResponse.Result)
	s.Require().Equal("{}", string(pingResponse.Result))
}

// TestStreamableTestSuite_Ping_InvalidSession tests ping request with invalid session ID
func (s *StreamableTestSuite) TestStreamableTestSuite_Ping_InvalidSession() {
	pingReq := NewJSONRPCRequest(s.T(), qilin.MethodPing, nil)
	pingReqBytes, err := json.Marshal(pingReq)
	s.Require().NoError(err)

	url := fmt.Sprintf("http://%s/mcp", s.address)
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(pingReqBytes))
	s.Require().NoError(err)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set(transport.MCPSessionID, "invalid-session-id")

	client := &http.Client{}
	pingResp, err := client.Do(httpReq)
	s.Require().NoError(err)
	defer pingResp.Body.Close()

	s.Require().Equal(http.StatusNotFound, pingResp.StatusCode)
}

// TestStreamableTestSuite_Ping_MissingSessionHeader tests ping request without session ID header
func (s *StreamableTestSuite) TestStreamableTestSuite_Ping_MissingSessionHeader() {
	pingReq := NewJSONRPCRequest(s.T(), qilin.MethodPing, nil)
	pingReqBytes, err := json.Marshal(pingReq)
	s.Require().NoError(err)

	url := fmt.Sprintf("http://%s/mcp", s.address)
	pingResp, err := http.Post(url, "application/json", bytes.NewReader(pingReqBytes))
	s.Require().NoError(err)
	defer pingResp.Body.Close()

	s.Require().Equal(http.StatusBadRequest, pingResp.StatusCode)
}

// TestStreamableTestSuite_PromptsList tests successful prompts/list request via HTTP
func (s *StreamableTestSuite) TestStreamableTestSuite_PromptsList() {
	// Initialize and get session ID
	sessionID := s.initializeSessionAndGetID()

	req := NewJSONRPCRequest(s.T(), qilin.MethodPromptsList, nil)
	reqBytes, err := json.Marshal(req)
	s.Require().NoError(err)

	url := fmt.Sprintf("http://%s/mcp", s.address)
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(reqBytes))
	s.Require().NoError(err)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set(transport.MCPSessionID, sessionID)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)

	responseData, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	response := JSONRPCResponseFromBytes(s.T(), responseData)

	s.Require().Equal(req.ID, response.ID)
	s.Require().Nil(response.Error)
	s.Require().NotNil(response.Result)

	var result map[string]any
	err = json.Unmarshal(response.Result, &result)
	s.Require().NoError(err)

	prompts, ok := result["prompts"].([]any)
	s.Require().True(ok)
	s.Require().Len(prompts, 1)

	prompt := prompts[0].(map[string]any)
	s.Require().Equal("greeting", prompt["name"])
	s.Require().Equal("A greeting prompt that welcomes users", prompt["description"])

	arguments, ok := prompt["arguments"].([]any)
	s.Require().True(ok)
	s.Require().Len(arguments, 1)

	arg := arguments[0].(map[string]any)
	s.Require().Equal("name", arg["name"])
	s.Require().Equal("The name of the person to greet", arg["description"])
	s.Require().Nil(arg["required"]) // Field is omitted when Required is false due to omitzero
}

// TestStreamableTestSuite_PromptsGet tests successful prompts/get request via HTTP
func (s *StreamableTestSuite) TestStreamableTestSuite_PromptsGet() {
	// Initialize and get session ID
	sessionID := s.initializeSessionAndGetID()

	params := map[string]any{
		"name": "greeting",
		"arguments": map[string]any{
			"name": "Alice",
		},
	}

	req := NewJSONRPCRequest(s.T(), qilin.MethodPromptsGet, params)
	reqBytes, err := json.Marshal(req)
	s.Require().NoError(err)

	url := fmt.Sprintf("http://%s/mcp", s.address)
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(reqBytes))
	s.Require().NoError(err)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set(transport.MCPSessionID, sessionID)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)

	responseData, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	response := JSONRPCResponseFromBytes(s.T(), responseData)

	s.Require().Equal(req.ID, response.ID)
	s.Require().Nil(response.Error)
	s.Require().NotNil(response.Result)

	var result map[string]any
	err = json.Unmarshal(response.Result, &result)
	s.Require().NoError(err)

	s.Require().Equal("A greeting prompt that welcomes users", result["description"])

	messages, ok := result["messages"].([]any)
	s.Require().True(ok)
	s.Require().Len(messages, 2)

	// Check system message
	systemMsg := messages[0].(map[string]any)
	s.Require().Equal("system", systemMsg["role"])
	systemContent := systemMsg["content"].(map[string]any)
	s.Require().Equal("text", systemContent["type"])
	s.Require().Equal("You are a helpful assistant.", systemContent["text"])

	// Check user message
	userMsg := messages[1].(map[string]any)
	s.Require().Equal("user", userMsg["role"])
	userContent := userMsg["content"].(map[string]any)
	s.Require().Equal("text", userContent["type"])
	s.Require().Equal("Hello, Alice! How can I help you today?", userContent["text"])
}

// TestStreamableTestSuite_PromptsGet_WithoutArguments tests prompts/get request without arguments via HTTP
func (s *StreamableTestSuite) TestStreamableTestSuite_PromptsGet_WithoutArguments() {
	// Initialize and get session ID
	sessionID := s.initializeSessionAndGetID()

	params := map[string]any{
		"name": "greeting",
	}

	req := NewJSONRPCRequest(s.T(), qilin.MethodPromptsGet, params)
	reqBytes, err := json.Marshal(req)
	s.Require().NoError(err)

	url := fmt.Sprintf("http://%s/mcp", s.address)
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(reqBytes))
	s.Require().NoError(err)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set(transport.MCPSessionID, sessionID)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)

	responseData, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)

	response := JSONRPCResponseFromBytes(s.T(), responseData)

	s.Require().Equal(req.ID, response.ID)
	s.Require().Nil(response.Error)
	s.Require().NotNil(response.Result)

	var result map[string]any
	err = json.Unmarshal(response.Result, &result)
	s.Require().NoError(err)

	messages, ok := result["messages"].([]any)
	s.Require().True(ok)
	s.Require().Len(messages, 2)

	// Check user message uses default name
	userMsg := messages[1].(map[string]any)
	userContent := userMsg["content"].(map[string]any)
	s.Require().Equal("Hello, World! How can I help you today?", userContent["text"])
}

// TestStreamableTestSuite_DeleteSession tests successful deletion of a session
func (s *StreamableTestSuite) TestStreamableTestSuite_DeleteSession() {
	params := map[string]any{
		"protocolVersion": qilin.LatestProtocolVersion,
		"capabilities": map[string]any{
			"experimental": map[string]any{},
		},
		"clientInfo": map[string]any{
			"name":    "test-client",
			"version": "1.0.0",
		},
	}
	initReq := NewJSONRPCRequest(s.T(), qilin.MethodInitialize, params)
	initReqBytes, err := json.Marshal(initReq)
	s.Require().NoError(err)

	url := fmt.Sprintf("http://%s/mcp", s.address)
	initResp, err := http.Post(url, "application/json", bytes.NewReader(initReqBytes))
	s.Require().NoError(err)
	defer initResp.Body.Close()

	sessionID := SessionIDFromResponse(s.T(), initResp)
	s.Require().NotEmpty(sessionID)

	httpReq, err := http.NewRequestWithContext(s.T().Context(), "DELETE", url, nil)
	s.Require().NoError(err)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set(transport.MCPSessionID, sessionID)

	client := &http.Client{}

	deleteResp, err := client.Do(httpReq)
	s.Require().NoError(err)
	defer deleteResp.Body.Close()

	s.Require().Equal(http.StatusNoContent, deleteResp.StatusCode)

	pingReq := NewJSONRPCRequest(s.T(), qilin.MethodPing, nil)
	pingReqBytes, err := json.Marshal(pingReq)
	s.Require().NoError(err)

	httpReq, err = http.NewRequest("POST", url, bytes.NewReader(pingReqBytes))
	s.Require().NoError(err)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set(transport.MCPSessionID, sessionID)

	pingResp, err := client.Do(httpReq)
	s.Require().NoError(err)
	defer pingResp.Body.Close()

	s.Require().Equal(http.StatusNotFound, pingResp.StatusCode)
}

// initializeSessionAndGetID helper function to initialize session and return session ID
func (s *StreamableTestSuite) initializeSessionAndGetID() string {
	params := map[string]any{
		"protocolVersion": qilin.LatestProtocolVersion,
		"capabilities": map[string]any{
			"experimental": map[string]any{},
		},
		"clientInfo": map[string]any{
			"name":    "test-client",
			"version": "1.0.0",
		},
	}

	req := NewJSONRPCRequest(s.T(), qilin.MethodInitialize, params)
	reqBytes, err := json.Marshal(req)
	s.Require().NoError(err)

	url := fmt.Sprintf("http://%s/mcp", s.address)
	resp, err := http.Post(url, "application/json", bytes.NewReader(reqBytes))
	s.Require().NoError(err)
	defer resp.Body.Close()

	sessionID := SessionIDFromResponse(s.T(), resp)
	s.Require().NotEmpty(sessionID)

	return sessionID
}
