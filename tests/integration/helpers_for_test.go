package integration

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"iter"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/miyamo2/qilin"
	"github.com/miyamo2/qilin/transport"
	"github.com/stretchr/testify/require"
)

// JSONRPCRequest represents a JSON-RPC request.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      string          `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// NewJSONRPCRequest creates a new JSON-RPC request with a random UUID as the ID.
func NewJSONRPCRequest(t *testing.T, method string, params any) JSONRPCRequest {
	t.Helper()
	_uuid, err := uuid.NewRandom()
	require.NoError(t, err, "failed to generate UUID for JSON-RPC request ID")

	p, err := json.Marshal(params)
	require.NoError(t, err, "failed to marshal JSON-RPC request parameters")

	return JSONRPCRequest{
		JSONRPC: qilin.JSONRPCVersion,
		ID:      _uuid.String(),
		Method:  method,
		Params:  p,
	}
}

// JSONRPCResponse represents a JSON-RPC response.
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      string          `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error.
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// StreamIterFromResponse reads the response body line by line and yields each line as a byte slice.
// It expects the response to have a Content-Type of "text/event-stream; charset=utf-8".
func StreamIterFromResponse(t *testing.T, resp *http.Response) iter.Seq[[]byte] {
	t.Helper()
	require.Equal(t, "text/event-stream; charset=utf-8", resp.Header.Get("Content-Type"))
	return func(yield func([]byte) bool) {
		defer resp.Body.Close()
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				require.ErrorIs(t, err, io.EOF)
			}
			if bytes.Contains(line, []byte("event: message")) {
				// Skip the "event: message" line
				continue
			}
			if bytes.HasPrefix(line, []byte("data: ")) {
				// Remove the "data: " prefix
				line = bytes.TrimPrefix(line, []byte("data: "))
			}
			if !yield(line) {
				break
			}
		}
	}
}

// SessionIDFromResponse extracts the session ID from the response headers.
func SessionIDFromResponse(t *testing.T, resp *http.Response) string {
	t.Helper()
	sessionID := resp.Header.Get(transport.MCPSessionID)
	require.NotEmpty(t, sessionID)
	return sessionID
}

// JSONRPCResponseFromBytes decodes the JSON-RPC response from the given byte slice.
func JSONRPCResponseFromBytes(t *testing.T, bytes []byte) JSONRPCResponse {
	t.Helper()
	var response JSONRPCResponse
	require.NoError(t, json.Unmarshal(bytes, &response))
	return response
}
