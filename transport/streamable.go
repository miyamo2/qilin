package transport

import (
	"bytes"
	"context"
	"fmt"
	internaltransport "github.com/miyamo2/qilin/internal/transport"
	"golang.org/x/exp/jsonrpc2"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// compatibility check
var (
	_ jsonrpc2.Listener = (*Streamable)(nil)
	_ jsonrpc2.Framer   = (*Streamable)(nil)
	_ jsonrpc2.Dialer   = (*Streamable)(nil)
)

// Streamable implements jsonrpc2.Listener, jsonrpc2.Framer, and jsonrpc2.Dialer
type Streamable struct {
	framer           jsonrpc2.Framer
	authorizer       Authorizer
	netListener      net.Listener
	rwc              chan *StreamableReadWriteCloser
	allowCORSOrigin  string
	allowCORSMethods string
	allowCORSHeaders string
	errCh            chan error
}

var (
	defaultAllowCORSOrigin  = []string{"*"}
	defaultAllowCORSMethods = []string{"POST", "GET", "OPTIONS", "DELETE"}
	defaultAllowCORSHeaders = []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"}
	defaultAccessControlAllowOrigin  = []string{"*"}
	defaultAccessControlAllowHeaders = []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"}
)

// Reader See: jsonrpc2.Framer#Reader
func (s *Streamable) Reader(rw io.Reader) jsonrpc2.Reader {
	return s.framer.Reader(rw)
}

// Writer See: jsonrpc2.Framer#Writer
func (s *Streamable) Writer(rw io.Writer) jsonrpc2.Writer {
	return s.framer.Writer(rw)
}

// Accept See: jsonrpc2.Listener#Accept
func (s *Streamable) Accept(ctx context.Context) (io.ReadWriteCloser, error) {
	for {
		select {
		case rwc := <-s.rwc:
			return internaltransport.NewQilinIO(rwc), nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// Close See: jsonrpc2.Listener#Close
func (s *Streamable) Close() error {
	defer close(s.rwc)
	err := s.netListener.Close()
	if err != nil {
		return err
	}
	err = <-s.errCh
	if err != nil {
		return err
	}
	return nil
}

// Dial See: jsonrpc2.Dialer#Dial
func (s *Streamable) Dial(ctx context.Context) (io.ReadWriteCloser, error) {
	select {
	case rwc := <-s.rwc:
		return internaltransport.NewQilinIO(rwc), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Dialer See: jsonrpc2.Listener#Dialer
func (s *Streamable) Dialer() jsonrpc2.Dialer {
	return s
}

func (s *Streamable) serveHTTP(w http.ResponseWriter, r *http.Request) {
	s.cors(w, r)
	w.Header().Set("content-type", "application/json; charset=utf-8")

	err := s.authorizer.Authorize(r.Header.Get("authorization"))
	if err != nil {
		slog.ErrorContext(r.Context(), "[qilin] authorize failed", "error", err)
		http.Error(w, "authorize failed", http.StatusUnauthorized)
		return
	}

	flusher := w.(http.Flusher)

	ctx, cancel := context.WithCancel(r.Context())
	rwc := &StreamableReadWriteCloser{
		w:             w,
		r:             r.Body,
		flusher:       flusher,
		requestHeader: r.Header,
		ctx:           ctx,
		cancel:        cancel,
	}
	context.AfterFunc(ctx, func() {
		rwc.Close()
	})
	s.rwc <- rwc
	<-ctx.Done()
}

func (s *Streamable) preflight(w http.ResponseWriter, r *http.Request) {
	s.cors(w, r)
}

func (s *Streamable) cors(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("access-control-allow-origin", s.allowCORSOrigin)
	w.Header().Set("access-control-allow-methods", s.allowCORSMethods)
	w.Header().Set("access-control-allow-headers", s.allowCORSHeaders)
}

type streamableOptions struct {
	address                         string
	netListener                     net.Listener
	accessControlAllowOrigin        []string
	accessControlAllowOriginMethods []string
	accessControlAllowOriginHeaders []string
	authorizer                      Authorizer
}

// StreamableOption configures the Streamable transport.
type StreamableOption func(*streamableOptions)

// StreamableWithAddress settings the address to listen on.
//
// If not set, it defaults to ":3001".
//
// If a net.Listener is provided, this option is ignored.
func StreamableWithAddress(address string) StreamableOption {
	return func(s *streamableOptions) {
		s.address = address
	}
}

// StreamableWithNetListener settings the net.Listener to use.
func StreamableWithNetListener(listener net.Listener) StreamableOption {
	return func(s *streamableOptions) {
		s.address = listener.Addr().String()
		s.netListener = listener
	}
}

// StreamableWithAccessControlAllowOrigin settings the allowed origins.
func StreamableWithAccessControlAllowOrigin(allowCORSOrigin []string) StreamableOption {
	return func(s *streamableOptions) {
		s.accessControlAllowOrigin = allowCORSOrigin
	}
}

// StreamableWithAccessControlAllowMethods settings the allowed CORS methods.
func StreamableWithAccessControlAllowMethods(allowCORSMethods []string) StreamableOption {
	return func(s *streamableOptions) {
		s.accessControlAllowOriginMethods = allowCORSMethods
	}
}

// StreamableWithAccessControlAllowHeaders settings the allowed CORS headers.
func StreamableWithAccessControlAllowHeaders(allowCORSHeaders []string) StreamableOption {
	return func(s *streamableOptions) {
		s.accessControlAllowOriginHeaders = allowCORSHeaders
	}
}

// StreamableWithAuthorizer settings the authorizer for the Streamable transport.
func StreamableWithAuthorizer(authorizer Authorizer) StreamableOption {
	return func(s *streamableOptions) {
		s.authorizer = authorizer
	}
}

// NewStreamable creates new Streamable transport.
func NewStreamable(options ...StreamableOption) *Streamable {
	opts := &streamableOptions{
		address:                         ":3001",
		accessControlAllowOrigin:        defaultAccessControlAllowOrigin,
		accessControlAllowOriginMethods: defaultAccessControlAllowMethods,
		accessControlAllowOriginHeaders: defaultAccessControlAllowHeaders,
		authorizer:                      DefaultAuthorizer(),
	}
	for _, opt := range options {
		opt(opts)
	}
	if opts.netListener == nil {
		var err error
		opts.netListener, err = net.Listen("tcp", opts.address)
		if err != nil {
			panic(err)
		}
	}

	s := &Streamable{
		netListener:      opts.netListener,
		rwc:              make(chan *StreamableReadWriteCloser),
		allowCORSOrigin:  strings.Join(opts.accessControlAllowOrigin, ","),
		allowCORSMethods: strings.Join(opts.accessControlAllowOriginMethods, ","),
		allowCORSHeaders: strings.Join(opts.accessControlAllowOriginHeaders, ","),
		errCh:            make(chan error, 1),
		framer:           newStreamableFramer(),
		authorizer:       opts.authorizer,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("OPTIONS /mcp", s.preflight)
	mux.HandleFunc("GET /mcp", s.serveHTTP)
	mux.HandleFunc("POST /mcp", s.serveHTTP)

	go func() {
		s.errCh <- http.Serve(s.netListener, mux)
	}()
	return s
}

// HttpIO wraps the http.ResponseWriter, http.Flusher and http.Request.Body with compatibility for the io.ReadWriteCloser
//
// WARNING: Do not implement this interface.
type HttpIO interface {
	SwitchStreamConnection(keepAlive time.Duration)
	Probe() error
	io.ReadWriteCloser
	SessionIDHolder
}

// compatibility check
var (
	_ io.ReadWriteCloser = (*StreamableReadWriteCloser)(nil)
	_ HttpIO             = (*StreamableReadWriteCloser)(nil)
)

type StreamableReadWriteCloser struct {
	_             struct{}
	w             http.ResponseWriter
	flusher       http.Flusher
	r             io.ReadCloser
	requestHeader http.Header
	ctx           context.Context
	cancel        context.CancelFunc
	sse           bool
	closeOnce     sync.Once
}

const (
	sseMessage = "event: message\ndata: %s\n\n"
)

// Read See: io.ReadWriteCloser#Read
func (s *StreamableReadWriteCloser) Read(p []byte) (n int, err error) {
	select {
	case <-s.ctx.Done():
		return 0, io.EOF
	default:
		// no-op
	}
	return s.r.Read(p)
}

// Write See: io.ReadWriteCloser#Write
func (s *StreamableReadWriteCloser) Write(p []byte) (n int, err error) {
	select {
	case <-s.ctx.Done():
		return 0, io.EOF
	default:
		// no-op
	}
	switch {
	case s.sse:
		n, err = fmt.Fprintf(s.w, sseMessage, p)
		if err != nil {
			return 0, err
		}
	default:
		defer s.Close()
		n, err = s.w.Write(p)
	}
	s.flusher.Flush()
	return
}

// Close See: io.ReadWriteCloser#Close
func (s *StreamableReadWriteCloser) Close() error {
	var err error
	s.closeOnce.Do(func() {
		s.r.Close()
		s.cancel()
		s.flusher.Flush()
		err = s.r.Close()
		if err != nil {
			return
		}
	})
	return err
}

// SessionID See: SessionIDHolder#SessionID
func (s *StreamableReadWriteCloser) SessionID() string {
	return s.requestHeader.Get(MCPSessionID)
}

// SetSessionID See: SessionIDHolder#SetSessionID
func (s *StreamableReadWriteCloser) SetSessionID(sessionID string) {
	s.w.Header().Set(MCPSessionID, sessionID)
}

// SwitchStreamConnection marks the StreamableReadWriteCloser as a streamable connection.
func (s *StreamableReadWriteCloser) SwitchStreamConnection(keepAlive time.Duration) {
	s.sse = true
	s.w.Header().Set("content-type", "text/event-stream; charset=utf-8")
	s.w.Header().Set("cache-control", "no-cache")
	s.w.Header().Set("connection", "keep-alive")
	s.w.Header().Set("keep-alive", fmt.Sprintf("timeout=%d", int(keepAlive.Seconds())))
	s.Probe()
	go func() {
		ticker := time.NewTicker(time.Duration(float64(keepAlive) * 0.8))
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.Probe()
			case <-s.ctx.Done():
				return
			}
		}
	}()
}

// Probe sends a comment message to the client to keep the connection alive.
func (s *StreamableReadWriteCloser) Probe() error {
	if _, err := s.w.Write([]byte(`:\n\n`)); err != nil {
		return fmt.Errorf("failed to write probe: %w", err)
	}
	s.flusher.Flush()
	return nil
}

// Context returns the context of the StreamableReadWriteCloser.
func (s *StreamableReadWriteCloser) Context() context.Context {
	//lint:ignore SA1029 Tentative hack to create a simple child context.
	ctx := context.WithValue(s.ctx, struct{}{}, struct{}{})
	return ctx
}

// compatibility check
var _ jsonrpc2.Writer = (*streamableWriter)(nil)

type streamableWriter struct {
	writerFunc func(rw io.Writer) jsonrpc2.Writer
	w          io.Writer
}

// Write See: jsonrpc2.Writer#Write
func (s *streamableWriter) Write(ctx context.Context, message jsonrpc2.Message) (int64, error) {
	var buf bytes.Buffer
	if _, err := s.writerFunc(&buf).Write(ctx, message); err != nil {
		return 0, err
	}
	if _, err := s.w.Write(buf.Bytes()); err != nil {
		return 0, err
	}
	return 0, nil
}

// newStreamableWriter returns a new streamable writer that uses the provided writer function.
func newStreamableWriter(w io.Writer, writerFunc func(rw io.Writer) jsonrpc2.Writer) jsonrpc2.Writer {
	return &streamableWriter{
		writerFunc: writerFunc,
		w:          w,
	}
}

// compatibility check
var _ jsonrpc2.Framer = (*streamableFramer)(nil)

// streamableFramer implements the jsonrpc2.Framer for streamable http transport.
type streamableFramer struct {
	jsonrpc2.Framer
}

// Writer implements the jsonrpc2.Framer#Writer
func (s *streamableFramer) Writer(w io.Writer) jsonrpc2.Writer {
	return newStreamableWriter(w, s.Framer.Writer)
}

// newStreamableFramer returns a new streamable framer.
func newStreamableFramer() jsonrpc2.Framer {
	return &streamableFramer{
		Framer: jsonrpc2.RawFramer(),
	}
}
