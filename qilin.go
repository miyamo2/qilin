package qilin

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/invopop/jsonschema"
	internaltransport "github.com/miyamo2/qilin/internal/transport"
	"github.com/miyamo2/qilin/transport"
	"golang.org/x/exp/jsonrpc2"
	"iter"
	"log/slog"
	"maps"
	"net/url"
	"reflect"
	"slices"
	"strings"
	"sync"
	"time"
	"weak"
)

// Qilin is the top-level framework instance.
type Qilin struct {
	// name of the server
	name string

	// version of the server
	version string

	// startupMutex is mutex to lock Qilin instance access during server configuration and startup.
	startupMutex sync.RWMutex

	// cold indicates the server is in a cold state. canceled when the server starts up.
	cold context.Context

	// warming is the cancel function to be called when the server starts up.
	warming context.CancelFunc

	// jsonUnmarshalFunc is the function to unmarshal JSON data
	jsonUnmarshalFunc JSONUnmarshalFunc

	// jsonMarshalFunc is the function to marshal JSON data
	jsonMarshalFunc JSONMarshalFunc

	// base64StringFunc is the function to encode binary data to a base64 string
	base64StringFunc Base64StringFunc

	// toolMiddleware is the list of toolMiddleware functions to be applied to each Tool handler
	toolMiddleware []ToolMiddlewareFunc

	// toolContextPool pools ToolContext
	toolContextPool sync.Pool

	// tools is the map of Tool names to Tool instances
	tools map[string]Tool

	// resourceMiddleware is the list of resourceMiddleware functions to be applied to each resource handler
	resourceMiddleware []ResourceMiddlewareFunc

	// resourceContextPool pools ResourceContext
	resourceContextPool sync.Pool

	// resources is the map of resource names to Resource instances
	resources map[string]Resource

	// resourceNode is the root node of the resource routing tree
	resourceNode resourceNode

	// resourceTemplates is the map of resource templates
	resourceTemplates map[string]resourceTemplate

	// resourceListHandler is the resource list handler
	resourceListHandler ResourceListHandlerFunc

	// resourceListContextPool pools ResourceListContext
	resourceListContextPool sync.Pool

	// resourceChangeSubscriberPool pools resourceChangeSubscriber
	resourceChangeSubscriberPool sync.Pool

	// resourceListChangeSubscriberPool pools resourceListChangeSubscriber
	resourceListChangeSubscriberPool sync.Pool

	resourceListChangeCtx *resourceListChangeContext

	// handlerPool pools handler
	handlerPool sync.Pool

	// capabilities is the map of capabilities
	capabilities ServerCapabilities

	// resourcesSubscriptionManager manage subscription status
	resourcesSubscriptionManager ResourcesSubscriptionManager

	// resourceSubscriptionOptions is the options for resource subscription
	resourcesSubscriptionOptions resourcesSubscriptionOptions

	// resourceListChangeSubscriptionOptions is the options for resource list subscription
	resourceListChangeSubscriptionOptions resourceListChangeSubscriptionOptions

	// resourceListChangeSubscriptionManager manage resource list subscription status
	resourceListChangeSubscriptionManager ResourceListChangeSubscriptionManager

	// sessionManagerOptions is the options for session management
	sessionManagerOptions sessionManagerOptions

	// sessionManager manage session
	sessionManager SessionManager

	// nowFunc is the function to get the current time
	nowFunc NowFunc
}

// resourcesSubscriptionOptions is the options for resource subscription.
type resourcesSubscriptionOptions struct {
	healthCheckInterval time.Duration
	store               ResourceModificationSubscriptionStore
}

// resourceListChangeSubscriptionOptions is the options for resource list subscription.
type resourceListChangeSubscriptionOptions struct {
	healthCheckInterval time.Duration
	store               ResourceListChangeSubscriptionStore
}

// sessionManagerOptions is the options for session management.
type sessionManagerOptions struct {
	store SessionStore
}

// ToolMiddlewareFunc defines a function to process Tool middleware.
type ToolMiddlewareFunc func(next ToolHandlerFunc) ToolHandlerFunc

// ToolHandlerFunc defines a function to serve Tool requests.
type ToolHandlerFunc func(c ToolContext) error

// ResourceHandlerFunc defines a function to serve resource requests.
type ResourceHandlerFunc func(c ResourceContext) error

// ResourceMiddlewareFunc defines a function to process resource middleware.
type ResourceMiddlewareFunc func(next ResourceHandlerFunc) ResourceHandlerFunc

// ResourceListHandlerFunc defines a function to serve resource list requests.
type ResourceListHandlerFunc func(c ResourceListContext) error

// DefaultResourceListHandler is the default resource list handler.
func DefaultResourceListHandler(c ResourceListContext) error {
	for k, v := range c.Resources() {
		paths := strings.Split(v.URI.Path, "/")
		if slices.ContainsFunc(paths, func(s string) bool {
			return strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")
		}) {
			continue
		}
		c.SetResource(k, v)
	}
	return nil
}

// ResourceListMiddlewareFunc defines a function to process resource list middleware.
type ResourceListMiddlewareFunc func(next ResourceListHandlerFunc) ResourceListHandlerFunc

// ResourceChangeObserverFunc defines a function to handle resource change notifications.
//
// The life cycle of this function must be in accordance with the application.
type ResourceChangeObserverFunc func(c ResourceChangeContext)

// ResourceListChangeObserverFunc defines a function to handle resource list change notifications
//
// The life cycle of this function must be in accordance with the application.
type ResourceListChangeObserverFunc func(c ResourceListChangeContext)

// JSONUnmarshalFunc defines a function to unmarshal JSON data.
type JSONUnmarshalFunc func(data []byte, v any) error

// JSONMarshalFunc defines a function to marshal JSON data.
type JSONMarshalFunc func(v any) ([]byte, error)

// Base64StringFunc defines a function to encode binary data to a base64 string.
type Base64StringFunc func(data []byte) string

// NowFunc defines a function to get the current time.
type NowFunc func() time.Time

// Option configures the Qilin instance.
type Option func(*Qilin)

func WithVersion(version string) Option {
	return func(q *Qilin) {
		q.version = version
	}
}

// WithJSONUnmarshalFunc sets the JSON unmarshal function.
func WithJSONUnmarshalFunc(f JSONUnmarshalFunc) Option {
	return func(q *Qilin) {
		q.jsonUnmarshalFunc = f
	}
}

// WithJSONMarshalFunc sets the JSON marshal function.
func WithJSONMarshalFunc(f JSONMarshalFunc) Option {
	return func(q *Qilin) {
		q.jsonMarshalFunc = f
	}
}

// WithNowFunc sets the function to get the current time.
func WithNowFunc(f NowFunc) Option {
	return func(q *Qilin) {
		q.nowFunc = f
	}
}

// WithResourceSubscriptionHealthCheckInterval sets the health check interval for the resource subscription.
func WithResourceSubscriptionHealthCheckInterval(interval time.Duration) Option {
	return func(q *Qilin) {
		q.resourcesSubscriptionOptions.healthCheckInterval = interval
	}
}

// WithResourceListChangeSubscriptionHealthCheckInterval sets the health check interval for the resource list subscription.
func WithResourceListChangeSubscriptionHealthCheckInterval(interval time.Duration) Option {
	return func(q *Qilin) {
		q.resourceListChangeSubscriptionOptions.healthCheckInterval = interval
	}
}

// WithResourcesListChangeSubscriptionStore sets the ResourceListChangeSubscriptionStore to the Qilin instance.
func WithResourcesListChangeSubscriptionStore(store ResourceListChangeSubscriptionStore) Option {
	return func(q *Qilin) {
		q.resourceListChangeSubscriptionOptions.store = store
	}
}

// WithResourcesSubscriptionStore sets the ResourceModificationSubscriptionStore to the Qilin instance.
func WithResourcesSubscriptionStore(store ResourceModificationSubscriptionStore) Option {
	return func(q *Qilin) {
		q.resourcesSubscriptionOptions.store = store
	}
}

// WithSessionStore sets the SessionStore to the Qilin instance.
func WithSessionStore(store SessionStore) Option {
	return func(q *Qilin) {
		q.sessionManagerOptions.store = store
	}
}

// New creates a new Qilin instance.
func New(name string, options ...Option) *Qilin {
	cold, warming := context.WithCancel(context.Background())
	q := &Qilin{
		name:              name,
		version:           "1.0.0",
		tools:             make(map[string]Tool),
		jsonMarshalFunc:   json.Marshal,
		jsonUnmarshalFunc: json.Unmarshal,
		base64StringFunc:  base64.StdEncoding.EncodeToString,
		resources:         make(map[string]Resource),
		resourceTemplates: make(map[string]resourceTemplate),
		resourceNode: resourceNode{
			child: resourceNodeChild(),
		},
		resourceListChangeCtx: &resourceListChangeContext{
			ctx:        context.Background(),
			subscriber: make(map[string]ResourceListChangeSubscriber),
		},
		cold:    cold,
		warming: warming,
		nowFunc: time.Now,
		resourceListChangeSubscriptionOptions: resourceListChangeSubscriptionOptions{
			healthCheckInterval: time.Minute,
		},
		resourcesSubscriptionOptions: resourcesSubscriptionOptions{
			healthCheckInterval: time.Minute,
		},
		sessionManagerOptions: sessionManagerOptions{
			store: &InMemorySessionStore{},
		},
	}
	ok := q.startupMutex.TryLock()
	if !ok {
		panic(ErrQilinLockingConflicts)
	}
	defer q.startupMutex.Unlock()

	q.resourceListHandler = DefaultResourceListHandler

	for _, opt := range options {
		opt(q)
	}
	q.toolContextPool = sync.Pool{
		New: func() any {
			return newToolContext(q.jsonUnmarshalFunc, q.jsonMarshalFunc, q.base64StringFunc)
		},
	}
	q.resourceContextPool = sync.Pool{
		New: func() any {
			return newResourceContext(q.jsonUnmarshalFunc, q.jsonMarshalFunc, q.base64StringFunc)
		},
	}
	q.resourceListContextPool = sync.Pool{
		New: func() any {
			return newResourceListContext(q.jsonUnmarshalFunc, q.jsonMarshalFunc)
		},
	}
	q.resourceChangeSubscriberPool = sync.Pool{
		New: func() any {
			return &resourceChangeSubscriber{
				nowFunc: q.nowFunc,
			}
		},
	}
	q.resourceListChangeSubscriberPool = sync.Pool{
		New: func() any {
			return &resourceListChangeSubscriber{
				nowFunc: q.nowFunc,
			}
		},
	}
	q.sessionManager = &sessionManager{
		store: q.sessionManagerOptions.store,
	}
	if q.resourceListChangeSubscriptionOptions.store == nil {
		q.resourceListChangeSubscriptionOptions.store = &InMemoryResourceListChangeSubscriptionStore{
			subscriptions:              sync.Map{},
			nowFunc:                    q.nowFunc,
			subscriptionHealthInterval: q.resourceListChangeSubscriptionOptions.healthCheckInterval,
		}
	}
	q.resourceListChangeSubscriptionManager = &resourceListChangeSubscriptionManager{
		nowFunc:                    q.nowFunc,
		subscriptionHealthInterval: q.resourceListChangeSubscriptionOptions.healthCheckInterval,
		store:                      q.resourceListChangeSubscriptionOptions.store,
	}
	if q.resourcesSubscriptionOptions.store == nil {
		q.resourcesSubscriptionOptions.store = &InMemoryResourceModificationSubscriptionStore{
			subscriptions:              sync.Map{},
			nowFunc:                    q.nowFunc,
			subscriptionHealthInterval: q.resourcesSubscriptionOptions.healthCheckInterval,
		}
	}
	q.resourcesSubscriptionManager = &resourcesSubscribeManager{
		nowFunc:                    q.nowFunc,
		subscriptionHealthInterval: q.resourcesSubscriptionOptions.healthCheckInterval,
		store:                      q.resourcesSubscriptionOptions.store,
	}
	return q
}

type toolOptions struct {
	description string
	annotation  ToolAnnotations
	middlewares []ToolMiddlewareFunc
}

// ToolOption configures the Tool options.
type ToolOption func(*toolOptions)

// ToolWithDescription configures the Tool description.
func ToolWithDescription(description string) ToolOption {
	return func(o *toolOptions) {
		o.description = description
	}
}

// ToolWithAnnotations configures the Tool annotations.
func ToolWithAnnotations(annotations ToolAnnotations) ToolOption {
	return func(o *toolOptions) {
		o.annotation = annotations
	}
}

// ToolWithMiddleware configures the Tool middleware.
func ToolWithMiddleware(middlewares ...ToolMiddlewareFunc) ToolOption {
	return func(o *toolOptions) {
		slices.Reverse(middlewares)
		o.middlewares = slices.Concat(middlewares, o.middlewares)
	}
}

// Tool registers a new Tool with the given name and description.
//
//   - name: the name of the Tool
//   - req: the request schema for the Tool
//   - handler: the handler function for the Tool
//   - options: (optional) the options for the Tool
func (q *Qilin) Tool(name string, req any, handler ToolHandlerFunc, options ...ToolOption) {
	ok := q.startupMutex.TryLock()
	if !ok {
		panic(ErrQilinLockingConflicts)
	}
	defer q.startupMutex.Unlock()

	if q.capabilities.Tools == nil {
		q.capabilities.Tools = &ToolCapability{}
	}

	opts := &toolOptions{}
	for _, o := range options {
		o(opts)
	}

	f := handler
	slices.Reverse(opts.middlewares)
	for _, m := range opts.middlewares {
		f = m(f)
	}
	ref := jsonschema.Reflector{
		Anonymous:      true,
		DoNotReference: true,
	}
	schema := ref.Reflect(req)
	schema.Version = ""
	q.tools[name] = Tool{
		Name:        name,
		Description: opts.description,
		InputSchema: schema,
		handler:     f,
	}
}

type resourceOptions struct {
	description string
	mimeType    string
	middlewares []ResourceMiddlewareFunc
}

// ResourceOption configures the resource options.
type ResourceOption func(*resourceOptions)

// ResourceWithDescription configures the resource description.
func ResourceWithDescription(description string) ResourceOption {
	return func(o *resourceOptions) {
		o.description = description
	}
}

// ResourceWithMimeType configures the resource MIME type.
func ResourceWithMimeType(mimeType string) ResourceOption {
	return func(o *resourceOptions) {
		o.mimeType = mimeType
	}
}

// ResourceWithMiddleware configures the resource middleware.
func ResourceWithMiddleware(middlewares ...ResourceMiddlewareFunc) ResourceOption {
	return func(o *resourceOptions) {
		slices.Reverse(middlewares)
		o.middlewares = slices.Concat(middlewares, o.middlewares)
	}
}

// Resource registers a new resource with the given name and description.
// If the URI contains path parameters, it will be registered as a template resource.
//
//   - name: the name of the resource
//   - uri: the URI of the resource
//   - handler: the handler function for the resource
//   - options: (optional) the options for the resource
func (q *Qilin) Resource(name, uri string, handler ResourceHandlerFunc, options ...ResourceOption) {
	ok := q.startupMutex.TryLock()
	if !ok {
		panic(ErrQilinLockingConflicts)
	}
	defer q.startupMutex.Unlock()

	if q.capabilities.Resources == nil {
		q.capabilities.Resources = &ResourceCapability{}
	}

	opts := &resourceOptions{}
	for _, o := range options {
		o(opts)
	}

	f := handler
	for _, m := range opts.middlewares {
		f = m(f)
	}
	resourceURI, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}
	if slices.ContainsFunc(strings.Split(resourceURI.Path, "/"), func(s string) bool {
		return strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")
	}) {
		q.resourceTemplates[resourceURI.Path] = resourceTemplate{
			URITemplate: (*ResourceURI)(resourceURI),
			Name:        name,
			Description: opts.description,
			MimeType:    opts.mimeType,
		}
	}
	n, _, _ := q.resourceNode.matching(resourceURI)
	if n != nil {
		n.handler = handler
		r := q.resources[resourceURI.String()]
		r.URI = (*ResourceURI)(resourceURI)
		r.Name = name
		r.Description = opts.description
		r.MimeType = opts.mimeType
		q.resources[resourceURI.String()] = r
		return
	}
	q.resourceNode.addRoute(resourceURI, handler, opts.mimeType)
	q.resources[resourceURI.String()] = Resource{
		URI:         (*ResourceURI)(resourceURI),
		Name:        name,
		Description: opts.description,
		MimeType:    opts.mimeType,
	}
}

// ResourceList registers a new resource list handler.
func (q *Qilin) ResourceList(handler ResourceListHandlerFunc, middleware ...ResourceListMiddlewareFunc) {
	ok := q.startupMutex.TryLock()
	if !ok {
		panic(ErrQilinLockingConflicts)
	}
	defer q.startupMutex.Unlock()
	f := handler
	slices.Reverse(middleware)
	for _, m := range middleware {
		f = m(f)
	}
	q.resourceListHandler = f
}

// ResourceChangeObserver registers a resource change observer for the given URI and runs the observer function.
func (q *Qilin) ResourceChangeObserver(uri string, observer ResourceChangeObserverFunc) {
	ok := q.startupMutex.TryLock()
	if !ok {
		panic(ErrQilinLockingConflicts)
	}
	defer q.startupMutex.Unlock()
	if q.capabilities.Resources == nil {
		q.capabilities.Resources = &ResourceCapability{}
	}
	if !q.capabilities.Resources.Subscribe {
		q.capabilities.Resources.Subscribe = true
	}
	resourceURI, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}

	n, _, _ := q.resourceNode.matching(resourceURI)
	resourceChangeCtx := &resourceChangeContext{
		ctx:        context.Background(),
		subscriber: make(map[string]ResourceChangeSubscriber),
	}
	if n != nil {
		n.resourceChangeCtx = resourceChangeCtx
	} else {
		q.resourceNode.addRoute(resourceURI, nil, "")
		q.resources[resourceURI.String()] = Resource{
			URI: (*ResourceURI)(resourceURI),
		}
	}
	q.handleResourceChangeObserver(observer, resourceChangeCtx)
}

// ResourceListChangeObserver registers a resource list change observer and runs the observer function.
func (q *Qilin) ResourceListChangeObserver(observer ResourceListChangeObserverFunc) {
	ok := q.startupMutex.TryLock()
	if !ok {
		panic(ErrQilinLockingConflicts)
	}
	defer q.startupMutex.Unlock()
	if q.capabilities.Resources == nil {
		q.capabilities.Resources = &ResourceCapability{}
	}
	if !q.capabilities.Resources.ListChanged {
		q.capabilities.Resources.ListChanged = true
	}
	q.handleResourceListChangeObserver(observer, q.resourceListChangeCtx)
}

// UseInTools adds middleware to the Tool handler chain.
func (q *Qilin) UseInTools(middleware ...ToolMiddlewareFunc) {
	ok := q.startupMutex.TryLock()
	if !ok {
		panic(ErrQilinLockingConflicts)
	}
	defer q.startupMutex.Unlock()
	slices.Reverse(middleware)
	q.toolMiddleware = slices.Concat(middleware, q.toolMiddleware)
}

// UseInResources adds middleware to the resource handler chain.
func (q *Qilin) UseInResources(middleware ...ResourceMiddlewareFunc) {
	ok := q.startupMutex.TryLock()
	if !ok {
		panic(ErrQilinLockingConflicts)
	}
	defer q.startupMutex.Unlock()
	slices.Reverse(middleware)
	q.resourceMiddleware = slices.Concat(middleware, q.resourceMiddleware)
}

func (q *Qilin) handleResourceChangeObserver(
	fn ResourceChangeObserverFunc,
	c ResourceChangeContext,
) {
	go func() {
		<-q.cold.Done()
		fn(c)
	}()
}

func (q *Qilin) handleResourceListChangeObserver(
	fn ResourceListChangeObserverFunc,
	c ResourceListChangeContext,
) {
	go func() {
		<-q.cold.Done()
		fn(c)
	}()
}

type startOptions struct {
	ctx       context.Context
	listener  jsonrpc2.Listener
	framer    jsonrpc2.Framer
	preempter jsonrpc2.Preempter
}

// StartOption configures the startup settings for the Qilin instance
type StartOption func(*startOptions)

// StartWithContext settings the context
func StartWithContext(ctx context.Context) StartOption {
	return func(o *startOptions) {
		o.ctx = ctx
	}
}

// StartWithListener settings the jsonrpc2.Listener
func StartWithListener[T *transport.Stdio | *transport.Streamable](listener T) StartOption {
	return func(o *startOptions) {
		switch v := any(listener).(type) {
		case *transport.Stdio:
			o.listener = v
			o.framer = transport.DefaultStdioFramer()
		case *transport.Streamable:
			o.listener = v
			o.framer = transport.DefaultStreamableFramer()
		}
	}
}

// StartWithFramer settings the jsonrpc2.Framer
func StartWithFramer(framer jsonrpc2.Framer) StartOption {
	return func(o *startOptions) {
		o.framer = framer
	}
}

// StartWithPreempter settings the jsonrpc2.Preempter
func StartWithPreempter(preempter jsonrpc2.Preempter) StartOption {
	return func(o *startOptions) {
		o.preempter = preempter
	}
}

// Start starts Qilin app
func (q *Qilin) Start(options ...StartOption) error {
	if !q.startupMutex.TryLock() {
		panic(ErrQilinLockingConflicts)
	}
	// Locked until the jsonrpc2 server is shut down.
	defer q.startupMutex.Unlock()

	o := &startOptions{
		ctx:    context.Background(),
		framer: transport.DefaultStdioFramer(),
	}
	for _, opt := range options {
		opt(o)
	}
	ctx, cancel := context.WithCancel(o.ctx)
	defer cancel()
	if o.listener == nil {
		o.listener = transport.NewStdio(ctx)
	}
	context.AfterFunc(ctx, func() {
		o.listener.Close()
	})

	for name, tool := range q.tools {
		for _, middleware := range q.toolMiddleware {
			tool.handler = middleware(tool.handler)
		}
		q.tools[name] = tool
	}

	for v := range q.resourceNode.flattenIter() {
		if v.handler != nil {
			continue
		}
		for _, middleware := range q.resourceMiddleware {
			v.handler = middleware(v.handler)
		}
	}
	var (
		enabledResourceListChange bool
		enabledResourceChange     bool
	)
	if q.capabilities.Resources != nil {
		enabledResourceListChange = q.capabilities.Resources.ListChanged
		enabledResourceChange = q.capabilities.Resources.Subscribe
	}
	q.handlerPool = sync.Pool{
		New: func() any {
			return &handler{
				qilin:                     q,
				enabledResourceListChange: enabledResourceListChange,
				enabledResourceChange:     enabledResourceChange,
				switchToStreamConnection:  noopFuncWithDuration,
				wg:                        sync.WaitGroup{},
			}
		},
	}

	srv, err := jsonrpc2.Serve(o.ctx, o.listener, newBinder(q, o.preempter, o.framer))
	if err != nil {
		return err
	}
	q.warming()
	return srv.Wait()
}

type Notify func(ctx context.Context, method string, params interface{}) error

// compatibility check
var _ jsonrpc2.Binder = (*binder)(nil)

type binder struct {
	qilin     *Qilin
	preempter jsonrpc2.Preempter
	framer    jsonrpc2.Framer
}

func (b *binder) Bind(_ context.Context, conn *jsonrpc2.Connection) (jsonrpc2.ConnectionOptions, error) {
	h := b.qilin.handlerPool.Get().(*handler)
	h.runningMu.Lock()
	h.notify = conn.Notify

	rv := reflect.ValueOf(conn).Elem()
	elem := rv.FieldByName("closer")

	qilinIO, ok := convertToQilinIO(elem)
	if !ok {
		return jsonrpc2.ConnectionOptions{}, errors.New("failed to convert to QilinIO")
	}
	switch inner := qilinIO.Inner.(type) {
	case *transport.Stdio:
		h.getSessionID = inner.SessionID
		h.setSessionID = inner.SetSessionID
		h.connectionCtx = inner.Context()
	case *transport.StreamableReadWriteCloser:
		h.getSessionID = inner.SessionID
		h.setSessionID = inner.SetSessionID
		h.switchToStreamConnection = inner.SwitchStreamConnection
		h.connectionCtx = inner.Context()
	}

	return jsonrpc2.ConnectionOptions{
		Preempter: b.preempter,
		Framer:    b.framer,
		Handler:   h,
	}, nil
}

func convertToQilinIO(elem reflect.Value) (_ *internaltransport.QilinIO, ok bool) {
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("[qilin] failed to convert to QilinIO", slog.Any("recover", rec))
		}
	}()
	rf := reflect.NewAt(elem.Type(), elem.Addr().UnsafePointer()).Elem()
	v, ok := rf.Interface().(*internaltransport.QilinIO)
	return v, ok
}

func newBinder(q *Qilin, preempter jsonrpc2.Preempter, framer jsonrpc2.Framer) *binder {
	return &binder{
		qilin:     q,
		preempter: preempter,
		framer:    framer,
	}
}

// compatibility check
var _ jsonrpc2.Handler = (*handler)(nil)

type handler struct {
	// qilin instance that is the parent of this handler
	qilin *Qilin

	// notify sends notifications to the client
	notify Notify

	// enabledResourceListChange indicates if resource list change is enabled
	enabledResourceListChange bool

	// enabledResourceChange indicates if resource change is enabled
	enabledResourceChange bool

	// getSessionID returns the session ID
	getSessionID func() string

	// setSessionID sets the session ID
	setSessionID func(string)

	// switchToStreamConnection switches the connection to a stream connection
	switchToStreamConnection func(keepAlive time.Duration)

	// connectionCtx is the context of the connection
	connectionCtx context.Context

	// runningMu is a mutex to protect the running state of the handler
	runningMu sync.Mutex

	// wg wait for all subscriptions to finish
	wg sync.WaitGroup
}

// Handle See: jsonrpc2.Handler.Handle
func (h *handler) Handle(ctx context.Context, req *jsonrpc2.Request) (interface{}, error) {
	var sessionID string

	defer func() {
		if req.Method != MethodInitialize {
			h.afterHandle(ctx, sessionID)
		}
		go h.reset()
	}()

	if req.Method == MethodInitialize {
		return h.handleInitialize(ctx, req, &sessionID)
	}

	sessionID = h.getSessionID()
	if sessionID == "" {
		return nil, jsonrpc2.ErrUnknown
	}
	return h.invokeMethod(ctx, req, sessionID)
}

func (h *handler) afterHandle(ctx context.Context, sessionID string) {
	if !h.enabledResourceListChange && !h.enabledResourceChange {
		return
	}
	unhealthySubscriptionUris, err := h.qilin.resourcesSubscriptionManager.UnhealthSubscriptions(ctx, sessionID)
	if err != nil {
		return
	}
	for _, v := range unhealthySubscriptionUris {
		h.setupResourceSubscription(ctx, sessionID, v)
	}

	health, err := h.qilin.resourceListChangeSubscriptionManager.Health(ctx, sessionID)
	if err != nil {
		return
	}
	if !health {
		sessionCtx, err := h.qilin.sessionManager.Context(ctx, sessionID)
		if err != nil {
			return
		}
		h.resourceListChangeSubscription(ctx, sessionCtx, sessionID)
	}
}

// invokeMethod invokes the method specified in the request.
func (h *handler) invokeMethod(ctx context.Context, req *jsonrpc2.Request, sessionID string) (interface{}, error) {
	sessionCtx, err := h.qilin.sessionManager.Context(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	select {
	case <-sessionCtx.Done():
	default:
		// no-op
	}
	switch req.Method {
	case MethodPing:
		return struct{}{}, nil
	case MethodResourcesList:
		return h.handleResourcesList(ctx, req)
	case MethodResourcesTemplatesList:
		return h.handleResourcesTemplatesList()
	case MethodResourcesRead:
		return h.handleResourcesRead(ctx, req)
	case MethodPromptsList, MethodPromptsGet:
		return nil, jsonrpc2.ErrNotHandled
	case MethodToolsList:
		return h.handleToolsList()
	case MethodToolsCall:
		return h.handleToolsCall(ctx, req)
	case MethodResourceSubscribe:
		if !h.enabledResourceChange {
			return nil, jsonrpc2.ErrMethodNotFound
		}
		return h.handleResourceSubscribe(ctx, sessionID, req)
	case MethodResourceUnsubscribe:
		return h.handleResourceUnsubscribe(ctx, sessionID, req)
	default:
		return nil, jsonrpc2.ErrMethodNotFound
	}
}

// handleInitialize handles the initialization request.
func (h *handler) handleInitialize(ctx context.Context, req *jsonrpc2.Request, sessionID *string) (interface{}, error) {
	var params initializeRequestParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, jsonrpc2.ErrInvalidParams
	}

	id, err := h.qilin.sessionManager.Start(ctx)
	if err != nil {
		return nil, err
	}
	h.setSessionID(id)
	*sessionID = id

	protocolVersion := params.ProtocolVersion
	if support := SupportedProtocolVersions[protocolVersion]; !support {
		protocolVersion = LatestProtocolVersion
	}

	sessionCtx, err := h.qilin.sessionManager.Context(ctx, *sessionID)
	if err != nil {
		return nil, err
	}

	if h.enabledResourceListChange {
		h.resourceListChangeSubscription(ctx, sessionCtx, *sessionID)
	}

	return &initializeResult{
		ProtocolVersion: protocolVersion,
		Capabilities:    h.qilin.capabilities,
		ServerInfo: implementation{
			Name:    h.qilin.name,
			Version: h.qilin.version,
		},
	}, nil
}

// resourceListChangeSubscription observes changes in the resource list and notifies the client.
func (h *handler) resourceListChangeSubscription(ctx context.Context, sessionCtx context.Context, sessionID string) error {
	listChangeCh := make(chan struct{}, 1)
	_resourceListChangeSubscriber := h.qilin.resourceListChangeSubscriberPool.Get().(*resourceListChangeSubscriber)
	_resourceListChangeSubscriber.id = sessionID
	_resourceListChangeSubscriber.ch = listChangeCh
	_resourceListChangeSubscriber.lastReceived = time.Now()
	h.qilin.resourceListChangeCtx.subscribe(_resourceListChangeSubscriber)

	subscription, err := h.qilin.resourceListChangeSubscriptionManager.SubscribeToResourceListChanges(ctx, sessionID)
	if err != nil {
		return err
	}
	h.switchToStreamConnection(5 * time.Second)
	h.wg.Add(1)

	go func() {
		defer h.wg.Done()
		defer h.qilin.resourceListChangeSubscriberPool.Put(_resourceListChangeSubscriber)
		defer _resourceListChangeSubscriber.reset()

		ticker := time.NewTicker(h.qilin.resourceListChangeSubscriptionOptions.healthCheckInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				subscription.SignalAlive()
			case <-sessionCtx.Done():
			case <-subscription.Unsubscribed():
			case <-h.connectionCtx.Done():
				return
			case <-listChangeCh:
				err = h.notify(h.connectionCtx, MethodNotificationResourcesListChanged, nil)
				if err != nil {
					return
				}
			}
		}
	}()
	return nil
}

// handleResourceList handles the request to list resources.
func (h *handler) handleResourcesList(ctx context.Context, req *jsonrpc2.Request) (interface{}, error) {
	dest := make(map[string]Resource)
	c := h.qilin.resourceListContextPool.Get().(*resourceListContext)
	c.ctx = ctx
	c.jsonrpcRequest = req
	c.dest = &dest
	c.resources = h.qilin.resources
	defer func() {
		c.reset()
		h.qilin.resourceListContextPool.Put(c)
	}()

	err := h.qilin.resourceListHandler(c)
	if err != nil {
		return nil, err
	}

	return &listResourcesResult{
		Resources: slices.Collect(maps.Values(dest)),
	}, nil
}

// handleResourcesTemplatesList handles the request to list resource templates.
func (h *handler) handleResourcesTemplatesList() (interface{}, error) {
	return &listResourceTemplatesResult{
		ResourceTemplates: slices.Collect(maps.Values(h.qilin.resourceTemplates)),
	}, nil
}

// handleResourcesRead handles the request to read a resource.
func (h *handler) handleResourcesRead(ctx context.Context, req *jsonrpc2.Request) (interface{}, error) {
	var params readResourceRequestParams
	if err := h.qilin.jsonUnmarshalFunc(req.Params, &params); err != nil {
		return nil, jsonrpc2.ErrInvalidParams
	}

	uri := (*url.URL)(params.URI)
	route, pathParam, err := h.qilin.resourceNode.matching(uri)
	if err != nil {
		return nil, err
	}

	c := h.qilin.resourceContextPool.Get().(*resourceContext)
	var dest readResourceResult
	c.ctx = ctx
	c.uri = weak.Make(uri)
	c.jsonrpcRequest = req
	c.pathParams = pathParam
	c.dest = &dest

	defer func() {
		c.reset()
		h.qilin.resourceContextPool.Put(c)
	}()

	err = route.handler(c)
	if err != nil {
		return nil, err
	}
	return &dest, nil
}

// handleToolsList handles the request to list tools.
func (h *handler) handleToolsList() (interface{}, error) {
	return &listToolsResponse{
		Tools: slices.Collect(maps.Values(h.qilin.tools)),
	}, nil
}

// handleToolsCall handles the request to call a tool.
func (h *handler) handleToolsCall(ctx context.Context, req *jsonrpc2.Request) (interface{}, error) {
	var params callToolRequestParams
	if err := h.qilin.jsonUnmarshalFunc(req.Params, &params); err != nil {
		return nil, jsonrpc2.ErrInvalidParams
	}

	tool, toolAvailable := h.qilin.tools[params.Name]
	if !toolAvailable {
		return nil, jsonrpc2.ErrInvalidParams
	}

	c := h.qilin.toolContextPool.Get().(*toolContext)
	var dest CallToolContent
	c.toolName = params.Name
	c.ctx = ctx
	c.jsonrpcRequest = req
	c.args = params.Arguments
	c.dest = &dest

	defer func() {
		c.reset()
		h.qilin.toolContextPool.Put(c)
	}()

	if err := tool.handler(c); err != nil {
		return nil, fmt.Errorf(ErrorMessageFailedToHandleTool, params.Name, err)
	}
	return dest, nil
}

// handleResourceSubscribe handles the request to subscribe to resource changes.
func (h *handler) handleResourceSubscribe(ctx context.Context, sessionID string, req *jsonrpc2.Request) (interface{}, error) {
	var params subscribeResourcesRequestParams
	if err := h.qilin.jsonUnmarshalFunc(req.Params, &params); err != nil {
		return nil, jsonrpc2.ErrInvalidParams
	}

	uri := (*url.URL)(params.URI)
	err := h.setupResourceSubscription(ctx, sessionID, uri)
	if err != nil {
		return nil, err
	}
	return struct{}{}, nil
}

// setupResourceSubscription sets up a subscription for resource changes.
func (h *handler) setupResourceSubscription(ctx context.Context, sessionID string, uri *url.URL) error {
	n, _, err := h.qilin.resourceNode.matching(uri)
	if err != nil {
		return err
	}

	resourceUpdateCh := make(chan *url.URL, 1)
	subscriber := h.qilin.resourceChangeSubscriberPool.Get().(*resourceChangeSubscriber)
	subscriber.ch = resourceUpdateCh
	subscriber.subscribedURI = uri
	subscriber.lastReceived = time.Now()
	subscriber.id = fmt.Sprintf("%s#%s", uri.String(), sessionID)

	n.resourceChangeCtx.subscribe(subscriber)
	subscription, err := h.qilin.resourcesSubscriptionManager.SubscribeToResourceModification(ctx, sessionID, uri)
	if err != nil {
		return err
	}

	h.resourceSubscription(ctx, n, subscriber, subscription, resourceUpdateCh)
	return nil
}

// resourceSubscription observes changes to a resource and notifies subscribers.
func (h *handler) resourceSubscription(
	ctx context.Context,
	n *resourceNode,
	subscriber *resourceChangeSubscriber,
	subscription Subscription,
	resourceUpdateCh chan *url.URL,
) {
	h.switchToStreamConnection(5 * time.Second)
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		ticker := time.NewTicker(h.qilin.resourcesSubscriptionOptions.healthCheckInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				subscription.SignalAlive()
			case <-ctx.Done():
			case <-subscription.Unsubscribed():
			case <-h.connectionCtx.Done():
				h.qilin.resourceChangeSubscriberPool.Put(subscriber)
				subscriber.reset()
				n.resourceChangeCtx.unsubscribe(subscriber.id)
				return
			case uri := <-resourceUpdateCh:
				if uri == nil {
					continue
				}
				err := h.notify(h.connectionCtx, MethodNotificationResourceUpdated, resourceUpdatedNotificationParam{
					URI: uri.String(),
				})
				if err != nil {
					return
				}
			}
		}
	}()

}

// handleResourceUnsubscribe handles the request to unsubscribe from resource changes.
func (h *handler) handleResourceUnsubscribe(ctx context.Context, sessionID string, req *jsonrpc2.Request) (interface{}, error) {
	var params unsubscribeResourcesRequestParams
	if err := h.qilin.jsonUnmarshalFunc(req.Params, &params); err != nil {
		return nil, jsonrpc2.ErrInvalidParams
	}

	uri := (*url.URL)(params.URI)
	err := h.qilin.resourcesSubscriptionManager.UnsubscribeToResourceModification(ctx, sessionID, uri)
	if err != nil {
		return nil, err
	}
	return struct{}{}, nil
}

func (h *handler) reset() {
	defer h.runningMu.Unlock()
	if h.runningMu.TryLock() {
		// if the handler is already reset, avoid double reset
		return
	}
	h.wg.Wait()
	h.notify = nil
	h.getSessionID = nil
	h.setSessionID = nil
	h.switchToStreamConnection = noopFuncWithDuration
	h.connectionCtx = nil
	h.qilin.handlerPool.Put(h)
}

type resourceNode struct {
	// child is the child resource node
	child *map[string]*resourceNode

	wild bool

	paramName string

	mimeType string

	// handler handles reading the resource.
	handler ResourceHandlerFunc

	// resourceChangeCtx can be used to subscribe to changes to this resource.
	resourceChangeCtx ResourceChangeContext
}

// matching finds the resource node that matches the given URI and parse the parameters
func (n *resourceNode) matching(uri *url.URL) (*resourceNode, map[string]string, error) {
	schema := uri.Scheme
	host := uri.Host
	path := strings.Split(strings.TrimPrefix(uri.Path, "/"), "/")

	params := make(map[string]string)
	child := *n.child
	r, ok := child[schema]
	if !ok {
		return nil, nil, fmt.Errorf("schema '%s' not found", schema)
	}
	child = *r.child
	r, ok = child[host]
	if !ok {
		return nil, nil, fmt.Errorf("host '%s' not found", host)
	}
	if len(path) == 0 {
		if r.handler != nil {
			return r, params, nil
		}
		return nil, nil, fmt.Errorf("host '%s' found, but not registered as a resource", path)
	}
	for _, p := range path {
		child = *r.child
		r, ok = child[p]
		if !ok {
			for _, v := range child {
				if v.wild {
					params[v.paramName] = p
					r = v
					break
				}
			}
			if r == nil {
				return nil, nil, fmt.Errorf("path '%s' not found", p)
			}
		}
	}
	if r.handler != nil {
		return r, params, nil
	}
	return nil, nil, fmt.Errorf("path '%s' found, but not registered as a resource", path)
}

// addRoute adds a new route to the resource node
func (n *resourceNode) addRoute(uri *url.URL, handler ResourceHandlerFunc, mimeType string) {
	schema := uri.Scheme
	host := uri.Host
	path := strings.Split(strings.TrimPrefix(uri.Path, "/"), "/")

	// Handle schema node
	schemaNode := n.getOrCreateChild(schema)

	// Handle host node
	hostNode := schemaNode.getOrCreateChild(host)

	// If path is empty, set handler on host node
	if len(path) == 0 {
		if handler != nil {
			hostNode.handler = handler
		}
		return
	}

	// Process path segments
	currentNode := hostNode
	for _, segment := range path {
		currentNode = currentNode.getOrCreateChild(segment)
	}

	// Set handler and mime type on final node
	if handler != nil {
		currentNode.handler = handler
	}
	if mimeType != "" {
		currentNode.mimeType = mimeType
	}
}

// getOrCreateChild gets or creates a child node for the given segment
func (n *resourceNode) getOrCreateChild(segment string) *resourceNode {
	child := *n.child
	node, ok := child[segment]
	if !ok {
		node = &resourceNode{
			child: resourceNodeChild(),
		}
		// Check if this is a path parameter (with curly braces)
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			node.wild = true
			node.paramName = strings.TrimPrefix(strings.TrimSuffix(segment, "}"), "{")
		}
		child[segment] = node
	}
	return node
}

// flattenIter flattens the resource node tree into a sequence of resource nodes
func (n *resourceNode) flattenIter() iter.Seq[*resourceNode] {
	return func(yield func(*resourceNode) bool) {
		if !yield(n) {
			return
		}
		for _, v := range *n.child {
			if len(*v.child) == 0 {
				if !yield(v) {
					return
				}
				continue
			}
			v.flattenIter()(yield)
		}
	}
}

// resourceNodeChild creates a new resource node child
func resourceNodeChild() *map[string]*resourceNode {
	v := make(map[string]*resourceNode)
	return &v
}

var (
	noopFuncWithDuration = func(_ time.Duration) {}
)
