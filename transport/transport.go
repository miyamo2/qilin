package transport

import (
	"golang.org/x/exp/jsonrpc2"
)

func DefaultFramer() jsonrpc2.Framer {
	return newStdioFramer()
}
