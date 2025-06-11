package integration

import (
	"context"
	"encoding/json"
	"io"
	"sync"
	"testing"

	"github.com/miyamo2/qilin"
	"github.com/miyamo2/qilin/transport"
	"github.com/stretchr/testify/suite"
)

type StdioTestSuite struct {
	suite.Suite
	mu              sync.Mutex
	clientReadPipe  io.ReadCloser
	clientWritePipe io.WriteCloser
	serverReadPipe  io.ReadCloser
	serverWritePipe io.WriteCloser
	cancel          context.CancelFunc
	ready           chan struct{}
}

func (s *StdioTestSuite) BeforeTest(_, _ string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create pipes for client-server communication
	s.serverReadPipe, s.clientWritePipe = io.Pipe()
	s.clientReadPipe, s.serverWritePipe = io.Pipe()
	s.ready = make(chan struct{})

	ctx, cancel := context.WithCancel(s.T().Context())
	s.cancel = cancel
	q := NewQilin(s.T())

	ready := make(chan struct{}, 1)
	go func() {
		q.Start(
			qilin.StartWithReadySignal(ready),
			qilin.StartWithContext(ctx),
			qilin.StartWithListener(
				transport.NewStdio(ctx,
					transport.StdioWithReadCloser(s.serverReadPipe),
					transport.StdioWithWriteCloser(s.serverWritePipe))))
	}()
	<-ready
}

func (s *StdioTestSuite) AfterTest(_, _ string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cancel()
	s.Require().NoError(s.clientWritePipe.Close())
	s.Require().NoError(s.clientReadPipe.Close())
	s.Require().NoError(s.serverReadPipe.Close())
	s.Require().NoError(s.serverWritePipe.Close())

	s.cancel = nil
	s.clientWritePipe = nil
	s.clientReadPipe = nil
	s.serverReadPipe = nil
	s.serverWritePipe = nil
}

func TestStdioTestSuite(t *testing.T) {
	suite.Run(t, new(StdioTestSuite))
}

// TestStdioTestSuite_Initialize_Success tests successful initialization process
func (s *StdioTestSuite) TestStdioTestSuite_Initialize_Success() {
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

	_, err = s.clientWritePipe.Write(append(reqBytes, '\n'))
	s.Require().NoError(err)

	buf := make([]byte, 4096)
	n, err := s.clientReadPipe.Read(buf)
	s.Require().NoError(err)

	response := JSONRPCResponseFromBytes(s.T(), buf[:n])

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
	s.Require().NotEmpty(version)
}

// TestStdioTestSuite_Initialize_ProtocolVersionFallback tests protocol version fallback
func (s *StdioTestSuite) TestStdioTestSuite_Initialize_ProtocolVersionFallback() {
	// Create initialization request with unsupported protocol version
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

	_, err = s.clientWritePipe.Write(append(reqBytes, '\n'))
	s.Require().NoError(err)

	buf := make([]byte, 4096)
	n, err := s.clientReadPipe.Read(buf)
	s.Require().NoError(err)

	response := JSONRPCResponseFromBytes(s.T(), buf[:n])

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
}

// TestStdioTestSuite_Ping_Success tests successful ping request after initialization
func (s *StdioTestSuite) TestStdioTestSuite_Ping_Success() {
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

	_, err = s.clientWritePipe.Write(append(initReqBytes, '\n'))
	s.Require().NoError(err)

	buf := make([]byte, 4096)
	n, err := s.clientReadPipe.Read(buf)
	s.Require().NoError(err)

	initResponse := JSONRPCResponseFromBytes(s.T(), buf[:n])
	s.Require().Equal(initReq.ID, initResponse.ID)
	s.Require().Nil(initResponse.Error)
	s.Require().NotNil(initResponse.Result)

	pingReq := NewJSONRPCRequest(s.T(), qilin.MethodPing, nil)
	pingReqBytes, err := json.Marshal(pingReq)
	s.Require().NoError(err)

	_, err = s.clientWritePipe.Write(append(pingReqBytes, '\n'))
	s.Require().NoError(err)

	buf = make([]byte, 4096)
	n, err = s.clientReadPipe.Read(buf)
	s.Require().NoError(err)

	pingResponse := JSONRPCResponseFromBytes(s.T(), buf[:n])

	s.Require().Equal(pingReq.ID, pingResponse.ID)
	s.Require().Nil(pingResponse.Error)
	s.Require().NotNil(pingResponse.Result)
	s.Require().Equal("{}", string(pingResponse.Result))
}

// TestStdioTestSuite_Ping_WithoutSession tests ping request without prior initialization
func (s *StdioTestSuite) TestStdioTestSuite_Ping_WithoutSession() {
	pingReq := NewJSONRPCRequest(s.T(), qilin.MethodPing, nil)
	pingReqBytes, err := json.Marshal(pingReq)
	s.Require().NoError(err)

	_, err = s.clientWritePipe.Write(append(pingReqBytes, '\n'))
	s.Require().NoError(err)

	buf := make([]byte, 4096)
	n, err := s.clientReadPipe.Read(buf)
	s.Require().NoError(err)

	pingResponse := JSONRPCResponseFromBytes(s.T(), buf[:n])

	s.Require().Equal(pingReq.ID, pingResponse.ID)
	s.Require().NotNil(pingResponse.Error)
	s.Require().Nil(pingResponse.Result)
	s.Require().Equal(-32001, pingResponse.Error.Code)
}

// TestStdioTestSuite_PromptsList tests successful prompts/list request
func (s *StdioTestSuite) TestStdioTestSuite_PromptsList() {
	// Initialize first
	s.initializeConnection()

	req := NewJSONRPCRequest(s.T(), qilin.MethodPromptsList, nil)
	reqBytes, err := json.Marshal(req)
	s.Require().NoError(err)

	_, err = s.clientWritePipe.Write(append(reqBytes, '\n'))
	s.Require().NoError(err)

	buf := make([]byte, 4096)
	n, err := s.clientReadPipe.Read(buf)
	s.Require().NoError(err)

	response := JSONRPCResponseFromBytes(s.T(), buf[:n])

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
	s.Require().Equal(false, arg["required"])
}

// TestStdioTestSuite_PromptsGet tests successful prompts/get request
func (s *StdioTestSuite) TestStdioTestSuite_PromptsGet() {
	// Initialize first
	s.initializeConnection()

	params := map[string]any{
		"name": "greeting",
		"arguments": map[string]any{
			"name": "Alice",
		},
	}

	req := NewJSONRPCRequest(s.T(), qilin.MethodPromptsGet, params)
	reqBytes, err := json.Marshal(req)
	s.Require().NoError(err)

	_, err = s.clientWritePipe.Write(append(reqBytes, '\n'))
	s.Require().NoError(err)

	buf := make([]byte, 4096)
	n, err := s.clientReadPipe.Read(buf)
	s.Require().NoError(err)

	response := JSONRPCResponseFromBytes(s.T(), buf[:n])

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

// TestStdioTestSuite_PromptsGet_WithoutArguments tests prompts/get request without arguments
func (s *StdioTestSuite) TestStdioTestSuite_PromptsGet_WithoutArguments() {
	// Initialize first
	s.initializeConnection()

	params := map[string]any{
		"name": "greeting",
	}

	req := NewJSONRPCRequest(s.T(), qilin.MethodPromptsGet, params)
	reqBytes, err := json.Marshal(req)
	s.Require().NoError(err)

	_, err = s.clientWritePipe.Write(append(reqBytes, '\n'))
	s.Require().NoError(err)

	buf := make([]byte, 4096)
	n, err := s.clientReadPipe.Read(buf)
	s.Require().NoError(err)

	response := JSONRPCResponseFromBytes(s.T(), buf[:n])

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

// initializeConnection helper function to initialize connection
func (s *StdioTestSuite) initializeConnection() {
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

	_, err = s.clientWritePipe.Write(append(reqBytes, '\n'))
	s.Require().NoError(err)

	buf := make([]byte, 4096)
	_, err = s.clientReadPipe.Read(buf)
	s.Require().NoError(err)
}
