package transport_test

import (
	"net"

	"github.com/miyamo2/qilin"
	"github.com/miyamo2/qilin/transport"
)

func ExampleNewStreamable() {
	streamable := transport.NewStreamable()
	q := qilin.New("example")
	q.Start(qilin.StartWithListener(streamable))
}

func ExampleNewStreamable_withNetListener() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	streamable := transport.NewStreamable(transport.StreamableWithNetListener(listener))
	q := qilin.New("example")
	q.Start(qilin.StartWithListener(streamable))
}

func ExampleNewStreamable_withAddress() {
	streamable := transport.NewStreamable(transport.StreamableWithAddress("127.0.0.1:0"))
	q := qilin.New("example")
	q.Start(qilin.StartWithListener(streamable))
}

func ExampleNewStreamable_withAuthorizer() {
	type authorizer struct {
		transport.Authorizer
	}
	streamable := transport.NewStreamable(transport.StreamableWithAuthorizer(&authorizer{}))

	q := qilin.New("example")
	q.Start(qilin.StartWithListener(streamable))
}

func ExampleNewStreamable_withAccessControlAllowOrigin() {
	streamable := transport.NewStreamable(
		transport.StreamableWithAccessControlAllowOrigin(
			[]string{"https://example.com", "https://*.example.com"},
		),
	)

	q := qilin.New("example")

	q.Start(qilin.StartWithListener(streamable))
}

func ExampleNewStreamable_withAccessControlAllowMethods() {
	streamable := transport.NewStreamable(
		transport.StreamableWithAccessControlAllowMethods([]string{"GET", "POST", "PUT", "DELETE"}))

	q := qilin.New("example")

	q.Start(qilin.StartWithListener(streamable))
}

func ExampleNewStreamable_withAccessControlAllowHeaders() {
	streamable := transport.NewStreamable(
		transport.StreamableWithAccessControlAllowHeaders(
			[]string{"Content-Type", "Authorization", transport.MCPSessionID},
		),
	)

	q := qilin.New("example")
	q.Start(qilin.StartWithListener(streamable))
}

func ExampleStreamableWithAuthorizer() {
	type authorizer struct {
		transport.Authorizer
	}
	streamable := transport.NewStreamable(
		transport.StreamableWithAuthorizer(&authorizer{}))

	q := qilin.New("example")

	q.Start(qilin.StartWithListener(streamable))
}

func ExampleStreamableWithAddress() {
	streamable := transport.NewStreamable(transport.StreamableWithAddress("127.0.0.1:0"))
	q := qilin.New("example")
	q.Start(qilin.StartWithListener(streamable))
}

func ExampleStreamableWithAccessControlAllowOrigin() {
	streamable := transport.NewStreamable(
		transport.StreamableWithAccessControlAllowOrigin(
			[]string{"https://example.com", "https://*.example.com"},
		),
	)

	q := qilin.New("example")

	q.Start(qilin.StartWithListener(streamable))
}

func ExampleStreamableWithAccessControlAllowMethods() {
	streamable := transport.NewStreamable(
		transport.StreamableWithAccessControlAllowMethods([]string{"GET", "POST", "PUT", "DELETE"}))

	q := qilin.New("example")

	q.Start(qilin.StartWithListener(streamable))
}

func ExampleStreamableWithAccessControlAllowHeaders() {
	streamable := transport.NewStreamable(
		transport.StreamableWithAccessControlAllowHeaders(
			[]string{"Content-Type", "Authorization", transport.MCPSessionID},
		),
	)

	q := qilin.New("example")

	q.Start(qilin.StartWithListener(streamable))
}
