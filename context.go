package qilin

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"golang.org/x/exp/jsonrpc2"
	"net/url"
	"strings"
	"sync"
	"time"
	"weak"
)

type Context interface {
	// Get retrieves data from the context.
	Get(key any) any
	// Set saves data in the context.
	Set(key any, val any)
	// JSONRPCRequest returns the JSONRPC request
	JSONRPCRequest() jsonrpc2.Request
	// Context returns the context
	Context() context.Context
	// SetContext sets the context
	SetContext(ctx context.Context)
}

var _ Context = (*_context)(nil)

type _context struct {
	ctx               context.Context
	store             sync.Map
	jsonrpcRequest    *jsonrpc2.Request
	jsonUnmarshalFunc JSONUnmarshalFunc
	jsonMarshalFunc   JSONMarshalFunc
}

func (c *_context) Get(key any) any {
	v, _ := c.store.Load(key)
	return v
}

func (c *_context) Set(key any, val any) {
	c.store.Store(key, val)
}

func (c *_context) JSONRPCRequest() jsonrpc2.Request {
	return *c.jsonrpcRequest
}

func (c *_context) Context() context.Context {
	return c.ctx
}

func (c *_context) SetContext(ctx context.Context) {
	c.ctx = ctx
}

func (c *_context) reset() {
	c.store.Clear()
	c.jsonrpcRequest = nil
	c.ctx = nil
}

// BindableContext is the context for handlers that able to bind JSON data
type BindableContext interface {
	Context
	// Bind binds json data into the provided type `i`.
	Bind(i any) error
}

// ToolContext is the context for Tool handlers
type ToolContext interface {
	BindableContext
	// ToolName returns the name of the Tool
	ToolName() string
	// Arguments return the arguments passed to the Tool
	Arguments() json.RawMessage
	// String sends plain text content
	String(s string) error
	// JSON sends JSON content
	JSON(i any) error
	// Image sends image content
	Image(data []byte, mimeType string) error
	// Audio sends audio content
	Audio(data []byte, mimeType string) error
	// JSONResource sends embed JSON resource content
	JSONResource(uri *url.URL, i any, mimeType string) error
	// StringResource sends embed string resource content
	StringResource(uri *url.URL, s string, mimeType string) error
	// BinaryResource sends embed binary resource content
	BinaryResource(uri *url.URL, data []byte, mimeType string) error
}

var (
	_ Context     = (*toolContext)(nil)
	_ ToolContext = (*toolContext)(nil)
)

type toolContext struct {
	_context
	toolName   string
	args       json.RawMessage
	ctx        context.Context
	annotation *ToolAnnotations
	dest       *CallToolContent
}

func (c *toolContext) Arguments() json.RawMessage {
	return c.args
}

func (c *toolContext) Bind(i any) error {
	args := c.Arguments()
	if len(args) == 0 {
		return nil
	}
	return c.jsonUnmarshalFunc(args, i)
}

func (c *toolContext) String(s string) error {
	*c.dest = &textCallToolContent{
		Text:        s,
		Annotations: c.annotation,
		marshal:     c.jsonMarshalFunc,
	}
	return nil
}

func (c *toolContext) JSON(i any) error {
	b, err := c.jsonMarshalFunc(i)
	if err != nil {
		return err
	}
	*c.dest = &textCallToolContent{
		Text:        string(b),
		Annotations: c.annotation,
		marshal:     c.jsonMarshalFunc,
	}
	return nil
}

func (c *toolContext) Image(data []byte, mimeType string) error {
	enc := base64.StdEncoding.EncodeToString(data)
	*c.dest = &imageCallToolContent{
		Data:     enc,
		MimeType: mimeType,
		marshal:  c.jsonMarshalFunc,
	}
	return nil
}

func (c *toolContext) Audio(data []byte, mimeType string) error {
	enc := base64.StdEncoding.EncodeToString(data)
	*c.dest = &audioCallToolContent{
		Data:     enc,
		MimeType: mimeType,
		marshal:  c.jsonMarshalFunc,
	}
	return nil
}

func (c *toolContext) JSONResource(uri *url.URL, i any, mimeType string) error {
	b, err := c.jsonMarshalFunc(i)
	if err != nil {
		return err
	}
	if mimeType == "" {
		mimeType = "application/json"
	}
	*c.dest = &embedResourceCallToolContent{
		Resource: &textResourceContent{
			resourceContentBase: resourceContentBase{
				uri:      weak.Make(uri),
				mimeType: mimeType,
			},
			text: string(b),
		},
		marshal: c.jsonMarshalFunc,
	}
	return nil
}

func (c *toolContext) StringResource(uri *url.URL, s string, mimeType string) error {
	if mimeType == "" {
		mimeType = "text/plain"
	}
	*c.dest = &embedResourceCallToolContent{
		Resource: &textResourceContent{
			resourceContentBase: resourceContentBase{
				uri:      weak.Make(uri),
				mimeType: mimeType,
			},
			text: s,
		},
		marshal: c.jsonMarshalFunc,
	}
	return nil
}

func (c *toolContext) BinaryResource(uri *url.URL, data []byte, mimeType string) error {
	enc := base64.StdEncoding.EncodeToString(data)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	*c.dest = &embedResourceCallToolContent{
		Resource: &binaryResourceContent{
			resourceContentBase: resourceContentBase{
				uri:      weak.Make(uri),
				mimeType: mimeType,
			},
			blob: enc,
		},
		marshal: c.jsonMarshalFunc,
	}
	return nil
}

func (c *toolContext) ToolName() string {
	return c.toolName
}

// reset resets the Tool context
func (c *toolContext) reset() {
	c._context.reset()
	c.annotation = nil
	c.toolName = ""
	c.dest = nil
	c.args = nil
}

// newToolContext creates a new Tool context
func newToolContext(jsonUnmarshalFunc JSONUnmarshalFunc, jsonMarshalFunc JSONMarshalFunc) *toolContext {
	return &toolContext{
		_context: _context{
			jsonUnmarshalFunc: jsonUnmarshalFunc,
			jsonMarshalFunc:   jsonMarshalFunc,
		},
	}
}

// ResourceContext is the context for resource handlers
type ResourceContext interface {
	Context
	// ResourceURI returns the uri of the resource
	ResourceURI() *url.URL
	// MimeType returns the mime type of the resource
	MimeType() string
	// Param retrieves the path parameter by name
	Param(name string) string
	// String sends plain text content
	String(s string) error
	// JSON sends JSON content
	JSON(i any) error
	// Blob sends blob content
	//
	//  - data: the blob data
	//  - mimeType: (Optional) the mime type of the blob. if not provided, a resource mime type will be used.
	Blob(data []byte, mimeType string) error
}

var _ ResourceContext = (*resourceContext)(nil)

type resourceContext struct {
	_context
	uri        weak.Pointer[url.URL]
	mimeType   string
	pathParams map[string]string
	dest       *readResourceResult
}

func (c *resourceContext) ResourceURI() *url.URL {
	return c.uri.Value()
}

func (c *resourceContext) MimeType() string {
	return c.mimeType
}

func (c *resourceContext) Param(name string) string {
	return c.pathParams[name]
}

func (c *resourceContext) String(s string) error {
	c.dest.Contents = append(c.dest.Contents, textResourceContent{
		resourceContentBase: resourceContentBase{
			uri:      c.uri,
			mimeType: "text/plain",
		},
		text:    s,
		marshal: c.jsonMarshalFunc,
	})
	return nil
}

func (c *resourceContext) JSON(i any) error {
	b, err := c.jsonMarshalFunc(i)
	if err != nil {
		return err
	}
	mimeType := c.mimeType
	if mimeType == "" {
		mimeType = "application/json"
	}
	c.dest.Contents = append(c.dest.Contents, textResourceContent{
		resourceContentBase: resourceContentBase{
			uri:      c.uri,
			mimeType: mimeType,
		},
		text:    string(b),
		marshal: c.jsonMarshalFunc,
	})
	return nil
}

func (c *resourceContext) Blob(data []byte, mimeType string) error {
	enc := base64.StdEncoding.EncodeToString(data)
	switch {
	case mimeType != "":
		// do nothing
	case c.mimeType != "":
		mimeType = c.mimeType
	default:
		mimeType = "application/octet-stream"
	}
	c.dest.Contents = append(c.dest.Contents, binaryResourceContent{
		resourceContentBase: resourceContentBase{
			uri:      c.uri,
			mimeType: mimeType,
		},
		blob:    enc,
		marshal: c.jsonMarshalFunc,
	})
	return nil
}

func (c *resourceContext) reset() {
	c._context.reset()
	c.uri = weak.Pointer[url.URL]{}
	c.mimeType = ""
	c.pathParams = nil
	c.dest = nil
}

// newResourceContext creates a new resource context
func newResourceContext(jsonUnmarshalFunc JSONUnmarshalFunc, jsonMarshalFunc JSONMarshalFunc) *resourceContext {
	return &resourceContext{
		_context: _context{
			jsonUnmarshalFunc: jsonUnmarshalFunc,
			jsonMarshalFunc:   jsonMarshalFunc,
		},
	}
}

// ResourceListContext is the context for resource list handlers
type ResourceListContext interface {
	Context

	// NextCursor

	// Resources return registered resources.
	//
	// This includes templates, which must be replaced with the actual available paths in SetResource.
	Resources() map[string]Resource

	// SetResource sets the resource by uri
	//
	// must be the actual path and resource, not the template.
	SetResource(uri string, resource Resource)
}

var _ ResourceListContext = (*resourceListContext)(nil)

type resourceListContext struct {
	_context
	resources map[string]Resource
	dest      *map[string]Resource
}

func (r *resourceListContext) Resources() map[string]Resource {
	return r.resources
}

func (r *resourceListContext) SetResource(uri string, resource Resource) {
	(*r.dest)[uri] = resource
}

func (r *resourceListContext) reset() {
	r._context.reset()
	r.resources = nil
	r.dest = nil
}

// newResourceListContext creates a new resource list context
func newResourceListContext(jsonUnmarshalFunc JSONUnmarshalFunc, jsonMarshalFunc JSONMarshalFunc) *resourceListContext {
	return &resourceListContext{
		_context: _context{
			jsonUnmarshalFunc: jsonUnmarshalFunc,
			jsonMarshalFunc:   jsonMarshalFunc,
		},
	}
}

// ResourceChangeSubscriber is the interface for resource change subscribers
type ResourceChangeSubscriber interface {
	// ID returns the unique ID of the subscriber
	ID() string
	// SubscribedURI returns the subscribed resource URI
	SubscribedURI() *url.URL
	// LastReceived returns the last received time of the subscriber
	LastReceived() time.Time
	// Publish publishes the resource change event
	Publish(uri *url.URL)
}

// compatibility check
var _ ResourceChangeSubscriber = (*resourceChangeSubscriber)(nil)

type resourceChangeSubscriber struct {
	id            string
	subscribedURI *url.URL
	lastReceived  time.Time
	ch            chan *url.URL
	mu            sync.RWMutex
}

func (r *resourceChangeSubscriber) ID() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.id
}

func (r *resourceChangeSubscriber) SubscribedURI() *url.URL {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.subscribedURI
}

func (r *resourceChangeSubscriber) LastReceived() time.Time {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.lastReceived
}

func (r *resourceChangeSubscriber) Publish(uri *url.URL) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastReceived = time.Now()
	r.ch <- uri
}

func (r *resourceChangeSubscriber) reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.id = ""
	r.lastReceived = time.Time{}
	if r.ch != nil {
		close(r.ch)
	}
	r.subscribedURI = nil
	r.ch = nil
}

// ResourceChangeContext is the context for resource change publish handlers.
type ResourceChangeContext interface {
	// Context returns the application scope context
	Context() context.Context

	// Publish publishes the resource change event
	Publish(uri *url.URL, modifiedAt time.Time)
	subscribe(subscriber ResourceChangeSubscriber)
	unsubscribe(id string)
}

// compatibility check
var _ ResourceChangeContext = (*resourceChangeContext)(nil)

type resourceChangeContext struct {
	ctx        context.Context
	mu         sync.RWMutex
	subscriber map[string]ResourceChangeSubscriber
}

func (r *resourceChangeContext) Context() context.Context {
	return r.ctx
}

func (r *resourceChangeContext) Publish(uri *url.URL, modifiedAt time.Time) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, subscriber := range r.subscriber {
		if !uriMatches(uri, subscriber.SubscribedURI()) {
			continue
		}
		if subscriber.LastReceived().After(modifiedAt) {
			continue
		}
		subscriber.Publish(uri)
	}
}

// uriMatches checks if the uri matches the subscribed URI
func uriMatches(uri *url.URL, subscribedURI *url.URL) bool {
	if uri == nil || subscribedURI == nil {
		return false
	}
	if subscribedURI.Scheme != uri.Scheme {
		return false
	}
	if subscribedURI.Host != uri.Host {
		return false
	}
	actualPath := strings.Split(uri.Path, "/")
	subscribedPath := strings.Split(subscribedURI.Path, "/")
	if len(actualPath) != len(subscribedPath) {
		return false
	}
	for i, v := range strings.Split(subscribedURI.Path, "/") {
		if strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}") {
			continue
		}
		if actualPath[i] != v {
			return false
		}
	}
	return true
}

func (r *resourceChangeContext) subscribe(subscriber ResourceChangeSubscriber) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.subscriber == nil {
		r.subscriber = make(map[string]ResourceChangeSubscriber)
	}
	r.subscriber[subscriber.ID()] = subscriber
}

func (r *resourceChangeContext) unsubscribe(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.subscriber, id)
}

// ResourceListChangeSubscriber sunscribe resource list changes
type ResourceListChangeSubscriber interface {
	// ID returns the unique ID of the subscriber
	ID() string
	// LastReceived returns the last received time of the subscriber
	LastReceived() time.Time
	// Publish publishes the resource list change event
	Publish()
}

// compatibility check
var _ ResourceListChangeSubscriber = (*resourceListChangeSubscriber)(nil)

type resourceListChangeSubscriber struct {
	id           string
	lastReceived time.Time
	ch           chan struct{}
	mu           sync.RWMutex
}

func (r *resourceListChangeSubscriber) ID() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.id
}

func (r *resourceListChangeSubscriber) LastReceived() time.Time {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.lastReceived
}

func (r *resourceListChangeSubscriber) Publish() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastReceived = time.Now()
	r.ch <- struct{}{}
}

func (r *resourceListChangeSubscriber) reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.id = ""
	r.lastReceived = time.Time{}
	if r.ch != nil {
		close(r.ch)
	}
	r.ch = nil
}

// ResourceListChangeContext is the context for resource list change publish handlers
type ResourceListChangeContext interface {
	// Context returns the application scope context
	Context() context.Context
	// Publish publishes the resource list change event
	Publish(modifiedAt time.Time)
}

// compatibility check
var _ ResourceListChangeContext = (*resourceListChangeContext)(nil)

type resourceListChangeContext struct {
	ctx        context.Context
	mu         sync.RWMutex
	subscriber map[string]ResourceListChangeSubscriber
}

func (r *resourceListChangeContext) Context() context.Context {
	return r.ctx
}

func (r *resourceListChangeContext) Publish(modifiedAt time.Time) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, subscriber := range r.subscriber {
		if subscriber.LastReceived().After(modifiedAt) {
			continue
		}
		subscriber.Publish()
	}
}

func (r *resourceListChangeContext) subscribe(subscriber ResourceListChangeSubscriber) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.subscriber == nil {
		r.subscriber = make(map[string]ResourceListChangeSubscriber)
	}
	r.subscriber[subscriber.ID()] = subscriber
}

func (r *resourceListChangeContext) unsubscribe(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.subscriber, id)
}
