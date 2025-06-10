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
	// Create valid initialization request
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

	// Verify response
	s.Require().Equal(req.ID, response.ID)
	s.Require().Nil(response.Error)
	s.Require().NotNil(response.Result)

	// Parse initialize result
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
