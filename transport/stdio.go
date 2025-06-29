package transport

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"sync"

	internaltransport "github.com/miyamo2/qilin/internal/transport"
	"golang.org/x/exp/jsonrpc2"
)

// compatibility check
var (
	_ jsonrpc2.Listener  = (*Stdio)(nil)
	_ jsonrpc2.Dialer    = (*Stdio)(nil)
	_ io.ReadWriteCloser = (*Stdio)(nil)
	_ SessionIDHolder    = (*Stdio)(nil)
	_ ErrorNotifier      = (*Stdio)(nil)
)

// Stdio implements the jsonrpc2.Listener, jsonrpc2.Dialer and io.ReadWriteCloser
type Stdio struct {
	_           struct{}
	in          io.ReadCloser
	out         io.WriteCloser
	closeOnce   sync.Once
	ctx         context.Context
	cancel      context.CancelFunc
	writeMu     sync.Mutex
	acceptMu    sync.Mutex
	sessionID   string
	sessionIDMu sync.RWMutex
}

// Accept implements the jsonrpc2.Listener#Accept
func (s *Stdio) Accept(ctx context.Context) (io.ReadWriteCloser, error) {
	select {
	case <-s.ctx.Done():
		_ = s.Close()
		return nil, io.EOF
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		s.acceptMu.Lock()
		return internaltransport.NewQilinIO(s), nil
	}
}

// Dial implements the jsonrpc2.Dialer#Dial
func (s *Stdio) Dial(_ context.Context) (io.ReadWriteCloser, error) {
	return internaltransport.NewQilinIO(s), nil
}

// Dialer implements the jsonrpc2.Listener#Dialer
func (s *Stdio) Dialer() jsonrpc2.Dialer {
	return s
}

// Read implements the io.ReadCloser#Read
func (s *Stdio) Read(p []byte) (n int, err error) {
	return s.in.Read(p)
}

// Write implements the io.ReadWriteCloser#Write
func (s *Stdio) Write(p []byte) (n int, err error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.out.Write(p)
}

// Close the stdin stream, stdout streams and the listener.
func (s *Stdio) Close() error {
	var err error
	s.closeOnce.Do(func() {
		s.cancel()
		err = s.in.Close()
		if err != nil {
			return
		}
		err = s.out.Close()
		if err != nil {
			return
		}
		err = s.ctx.Err()
	})
	return err
}

// SessionID See: SessionIDHolder#SessionID
func (s *Stdio) SessionID() string {
	s.sessionIDMu.RLock()
	defer s.sessionIDMu.RUnlock()
	return s.sessionID
}

// SetSessionID See: SessionIDHolder#SetSessionID
func (s *Stdio) SetSessionID(sessionID string) {
	s.sessionIDMu.Lock()
	defer s.sessionIDMu.Unlock()
	if s.sessionID == "" {
		s.sessionID = sessionID
	}
}

// Context returns the context of the Stdio listener.
func (s *Stdio) Context() context.Context {
	//nolint:staticcheck
	//lint:ignore SA1029 Tentative hack to create a simple child context.
	ctx := context.WithValue(s.ctx, struct{}{}, struct{}{})
	return ctx
}

func (s *Stdio) NoticeError(err error) {
	slog.Error("[qilin] occurred an error in stdio transport", slog.String("error", err.Error()))
}

type stdioOptions struct {
	in  io.ReadCloser
	out io.WriteCloser
}

// StdioOption options for the Stdio listener.
type StdioOption func(*stdioOptions)

// StdioWithReadCloser setting the input stream for the Stdio listener.
// mainly used for testing purposes.
func StdioWithReadCloser(in io.ReadCloser) StdioOption {
	return func(o *stdioOptions) {
		o.in = in
	}
}

// StdioWithWriteCloser setting the output stream for the Stdio listener.
func StdioWithWriteCloser(out io.WriteCloser) StdioOption {
	return func(o *stdioOptions) {
		o.out = out
	}
}

// NewStdio returns a new Stdio listener.
func NewStdio(ctx context.Context, options ...StdioOption) *Stdio {
	opts := &stdioOptions{
		in:  os.Stdin,
		out: os.Stdout,
	}
	for _, opt := range options {
		opt(opts)
	}
	ctx, cancel := context.WithCancel(ctx)
	s := &Stdio{
		in:     opts.in,
		out:    opts.out,
		ctx:    ctx,
		cancel: cancel,
	}
	context.AfterFunc(ctx, func() {
		_ = s.Close()
	})
	return s
}

// compatibility check
var _ jsonrpc2.Framer = (*stdioFramer)(nil)

// stdioFramer implements the jsonrpc2.Framer for stdio transport.
type stdioFramer struct {
	jsonrpc2.Framer
}

// Writer implements the jsonrpc2.Framer#Writer
func (s stdioFramer) Writer(w io.Writer) jsonrpc2.Writer {
	return newStdioWriter(w, s.Framer.Writer)
}

// newStdioFramer returns a new stdio framer.
func newStdioFramer() jsonrpc2.Framer {
	return &stdioFramer{
		Framer: jsonrpc2.RawFramer(),
	}
}

// compatibility check
var _ jsonrpc2.Writer = (*stdioWriter)(nil)

// stdioWriter implements the jsonrpc2.Writer for stdio transport.
type stdioWriter struct {
	writerFunc func(rw io.Writer) jsonrpc2.Writer
	w          io.Writer
}

var (
	// stdioDelimiter is the delimiter used to separate messages in stdio transport.
	stdioDelimiter      byte = '\n'
	stdioDelimiterBytes      = []byte{stdioDelimiter}
)

// Write See: jsonrpc2.Writer#Write
func (s *stdioWriter) Write(ctx context.Context, message jsonrpc2.Message) (int64, error) {
	var buf bytes.Buffer
	if _, err := s.writerFunc(&buf).Write(ctx, message); err != nil {
		return 0, err
	}
	if _, err := buf.Write(stdioDelimiterBytes); err != nil {
		return 0, err
	}
	if _, err := s.w.Write(buf.Bytes()); err != nil {
		return 0, err
	}
	return 0, nil
}

// newStdioWriter returns a new stdio writer.
func newStdioWriter(w io.Writer, writerFunc func(rw io.Writer) jsonrpc2.Writer) jsonrpc2.Writer {
	return &stdioWriter{
		writerFunc: writerFunc,
		w:          w,
	}
}
