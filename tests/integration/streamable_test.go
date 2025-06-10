package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
