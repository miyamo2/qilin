package qilin

import (
	"context"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/invopop/jsonschema"
	"github.com/miyamo2/qilin/transport"
	"github.com/oklog/ulid/v2"
	"golang.org/x/exp/jsonrpc2"
	"iter"
	"maps"
	"net/url"
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

	// cold indicates the server is in cold state. cancelled when the server starts up.
	cold context.Context

	// uncold is the cancel function to be called when the server starts up.
	uncold context.CancelFunc

	// jsonUnmarshalFunc is the function to unmarshal JSON data
	jsonUnmarshalFunc JSONUnmarshalFunc

	// jsonMarshalFunc is the function to marshal JSON data
	jsonMarshalFunc JSONMarshalFunc

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

	// capabilities is the map of capabilities
	capabilities ServerCapabilities
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

// New creates a new Qilin instance.
func New(name string, options ...Option) *Qilin {
	cold, uncold := context.WithCancel(context.Background())
	q := &Qilin{
		name:              name,
		version:           "1.0.0",
		tools:             make(map[string]Tool),
		jsonMarshalFunc:   json.Marshal,
		jsonUnmarshalFunc: json.Unmarshal,
		resources:         make(map[string]Resource),
		resourceTemplates: make(map[string]resourceTemplate),
		resourceNode: resourceNode{
			child: resourceNodeChild(),
		},
		resourceListChangeCtx: &resourceListChangeContext{
			ctx:        context.Background(),
			subscriber: make(map[string]ResourceListChangeSubscriber),
		},
		cold:   cold,
		uncold: uncold,
	}
	ok := q.startupMutex.TryLock()
	if !ok {
		panic(ErrQilinLockingConflicts)
	}
	defer q.startupMutex.Unlock()

	q.resourceListHandler = func(c ResourceListContext) error {
		for k, v := range q.resources {
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
	for _, opt := range options {
		opt(q)
	}
	q.toolContextPool = sync.Pool{
		New: func() any {
			return newToolContext(q.jsonUnmarshalFunc, q.jsonMarshalFunc)
		},
	}
	q.resourceContextPool = sync.Pool{
		New: func() any {
			return newResourceContext(q.jsonUnmarshalFunc, q.jsonMarshalFunc)
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
				mu: sync.RWMutex{},
			}
		},
	}
	q.resourceListChangeSubscriberPool = sync.Pool{
		New: func() any {
			return &resourceListChangeSubscriber{
				mu: sync.RWMutex{},
			}
		},
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
	q.resourceNode.addRoute(resourceURI, handler)
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
		q.resourceNode.addRoute(resourceURI, nil)
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

type Notify func(ctx context.Context, method string, params interface{}) error

// handler returns jsonrpc2.HandlerFunc
func (q *Qilin) handler(rootCtx context.Context, notify Notify, connectionClosed <-chan struct{}) jsonrpc2.HandlerFunc {
	subscribedResources := sync.Map{}
	listChangeCh := make(chan struct{}, 1)
	resourceUpdateCh := make(chan *url.URL, 1)
	var _resourceListChangeSubscriber *resourceListChangeSubscriber

	go func() {
		defer close(listChangeCh)
		defer close(resourceUpdateCh)
		defer func() {
			subscribedResources.Range(func(key, value interface{}) bool {
				subscriber := value.(*resourceChangeSubscriber)
				n, _, _ := q.resourceNode.matching(subscriber.SubscribedURI())
				if n != nil {
					n.resourceChangeCtx.unsubscribe(subscriber.id)
				}
				subscriber.reset()
				q.resourceChangeSubscriberPool.Put(subscriber)
				return true
			})
			if _resourceListChangeSubscriber != nil {
				q.resourceListChangeCtx.unsubscribe(_resourceListChangeSubscriber.id)
				_resourceListChangeSubscriber.reset()
				q.resourceListChangeSubscriberPool.Put(_resourceListChangeSubscriber)
			}
		}()
		for {
			select {
			case <-listChangeCh:
				notify(rootCtx, MethodNotificationResourcesListChanged, nil)
			case uri := <-resourceUpdateCh:
				if uri == nil {
					continue
				}
				notify(rootCtx, MethodNotificationResourceUpdated, resourceUpdatedNotificationParam{
					URI: uri.String(),
				})
			case <-rootCtx.Done():
				return
			case <-connectionClosed:
				return
			}
		}
	}()
	return func(ctx context.Context, req *jsonrpc2.Request) (interface{}, error) {
		switch req.Method {
		case MethodInitialize:
			var params initializeRequestParams
			if err := json.Unmarshal(req.Params, &params); err != nil {
				return nil, jsonrpc2.ErrInvalidParams
			}
			protocolVersion := params.ProtocolVersion
			if support := SupportedProtocolVersions[protocolVersion]; !support {
				protocolVersion = LatestProtocolVersion
			}
			if q.capabilities.Resources != nil {
				if q.capabilities.Resources.ListChanged {
					_resourceListChangeSubscriber = q.resourceListChangeSubscriberPool.Get().(*resourceListChangeSubscriber)
					_resourceListChangeSubscriber.id = ulid.Make().String()
					_resourceListChangeSubscriber.ch = listChangeCh
					_resourceListChangeSubscriber.lastReceived = time.Now()
					q.resourceListChangeCtx.subscribe(_resourceListChangeSubscriber)
				}
			}
			return &initializeResult{
				ProtocolVersion: protocolVersion,
				Capabilities:    q.capabilities,
				ServerInfo: implementation{
					Name:    q.name,
					Version: q.version,
				},
			}, nil
		case MethodPing:
			return struct{}{}, nil
		case MethodResourcesList:
			dest := make(map[string]Resource)
			c := q.resourceListContextPool.Get().(*resourceListContext)

			c.ctx = ctx
			c.jsonrpcRequest = req
			c.dest = &dest
			c.resources = q.resources
			defer func() {
				c.reset()
				q.resourceListContextPool.Put(c)
			}()

			err := q.resourceListHandler(c)
			if err != nil {
				return nil, err
			}
			return &listResourcesResult{
				Resources: slices.Collect(maps.Values(dest)),
			}, nil
		case MethodResourcesTemplatesList:
			return &listResourceTemplatesResult{
				ResourceTemplates: slices.Collect(maps.Values(q.resourceTemplates)),
			}, nil
		case MethodResourcesRead:
			var params readResourceRequestParams
			if err := q.jsonUnmarshalFunc(req.Params, &params); err != nil {
				return nil, jsonrpc2.ErrInvalidParams
			}

			uri := (*url.URL)(params.URI)
			route, pathParam, err := q.resourceNode.matching(uri)
			if err != nil {
				return nil, err
			}
			c := q.resourceContextPool.Get().(*resourceContext)
			if err != nil {
				return nil, err
			}
			var dest readResourceResult
			c.ctx = ctx
			c.uri = weak.Make(uri)
			c.jsonrpcRequest = req
			c.pathParams = pathParam
			c.dest = &dest

			defer func() {
				c.reset()
				q.resourceContextPool.Put(c)
			}()

			err = route.handler(c)
			if err != nil {
				return nil, err
			}
			return &dest, nil
		case MethodPromptsList:
			// not yet implemented
			return nil, jsonrpc2.ErrNotHandled
		case MethodPromptsGet:
			// not yet implemented
			return nil, jsonrpc2.ErrNotHandled
		case MethodToolsList:
			return &listToolsResponse{
				Tools: slices.Collect(maps.Values(q.tools)),
			}, nil
		case MethodToolsCall:
			var params callToolRequestParams
			if err := q.jsonUnmarshalFunc(req.Params, &params); err != nil {
				return nil, jsonrpc2.ErrInvalidParams
			}
			tool, toolAvailable := q.tools[params.Name]
			if !toolAvailable {
				return nil, jsonrpc2.ErrInvalidParams
			}
			c := q.toolContextPool.Get().(*toolContext)

			var dest CallToolContent

			c.toolName = params.Name
			c.ctx = ctx
			c.jsonrpcRequest = req
			c.args = params.Arguments
			c.dest = &dest

			defer func() {
				c.reset()
				q.toolContextPool.Put(c)
			}()

			if err := tool.handler(c); err != nil {
				return nil, fmt.Errorf(ErrorMessageFailedToHandleTool, params.Name, err)
			}
			return dest, nil
		case MethodResourceSubscribe:
			var params subscribeResourcesRequestParams
			if err := q.jsonUnmarshalFunc(req.Params, &params); err != nil {
				return nil, jsonrpc2.ErrInvalidParams
			}
			uri := (*url.URL)(params.URI)
			n, _, err := q.resourceNode.matching(uri)
			if err != nil {
				return nil, err
			}

			subscriber := q.resourceChangeSubscriberPool.Get().(*resourceChangeSubscriber)
			subscribedResources.Store(uri.String(), subscriber)

			subscriber.ch = resourceUpdateCh
			subscriber.subscribedURI = uri
			subscriber.lastReceived = time.Now()
			subscriber.id = ulid.Make().String()

			n.resourceChangeCtx.subscribe(subscriber)
			return struct{}{}, nil
		case MethodResourceUnsubscribe:
			var params unsubscribeResourcesRequestParams
			if err := q.jsonUnmarshalFunc(req.Params, &params); err != nil {
				return nil, jsonrpc2.ErrInvalidParams
			}
			uri := (*url.URL)(params.URI)
			n, _, err := q.resourceNode.matching(uri)
			if err != nil {
				return nil, err
			}
			v, ok := subscribedResources.LoadAndDelete(uri.String())
			if !ok {
				return struct{}{}, nil
			}
			subscriber := v.(*resourceChangeSubscriber)
			n.resourceChangeCtx.unsubscribe(subscriber.id)
			subscriber.reset()
			q.resourceChangeSubscriberPool.Put(subscriber)
			return struct{}{}, nil
		default:
			return nil, jsonrpc2.ErrMethodNotFound
		}
	}
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
	ok := q.startupMutex.TryLock()
	if !ok {
		panic(ErrQilinLockingConflicts)
	}
	// Locked until the jsonrpc2 server is shut down.
	defer q.startupMutex.Unlock()

	o := &startOptions{
		ctx:    context.Background(),
		framer: transport.DefaultFramer(),
	}
	for _, opt := range options {
		opt(o)
	}
	connectionClosed := make(chan struct{})
	ctx, cancel := context.WithCancel(o.ctx)
	defer cancel()
	context.AfterFunc(ctx, func() {
		close(connectionClosed)
	})
	listener := transport.NewStdio(ctx, cancel)

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
	var binder binderFunc = func(ctx context.Context, conn *jsonrpc2.Connection) (jsonrpc2.ConnectionOptions, error) {
		handler := q.handler(o.ctx, conn.Notify, connectionClosed)
		return jsonrpc2.ConnectionOptions{
			Preempter: o.preempter,
			Framer:    o.framer,
			Handler:   handler,
		}, nil
	}

	srv, err := jsonrpc2.Serve(o.ctx, listener, binder)
	if err != nil {
		return err
	}
	q.uncold()
	return srv.Wait()
}

type resourceNode struct {
	// child is the child resource node
	child *map[string]*resourceNode

	wild bool

	paramName string

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
func (n *resourceNode) addRoute(uri *url.URL, handler ResourceHandlerFunc) {
	schema := uri.Scheme
	host := uri.Host
	path := strings.Split(strings.TrimPrefix(uri.Path, "/"), "/")

	child := *n.child
	r, ok := child[schema]
	if !ok {
		r = &resourceNode{
			child: resourceNodeChild(),
		}
		child[schema] = r
	}
	child = *r.child
	r, ok = child[host]
	if !ok {
		r = &resourceNode{
			child: resourceNodeChild(),
		}
		child[host] = r
	}
	if len(path) == 0 {
		if handler != nil {
			r.handler = handler
		}
		return
	}

	p := ""
	for _, p = range path {
		child = *r.child
		r, ok = child[p]
		if !ok {
			r = &resourceNode{
				child: resourceNodeChild(),
			}
			if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") {
				r.wild = true
				r.paramName = strings.TrimPrefix(strings.TrimSuffix(p, "}"), "{")
			}
			child[p] = r
		}
	}
	if handler != nil {
		r.handler = handler
	}
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

var _ jsonrpc2.Binder = (*binderFunc)(nil)

type binderFunc func(ctx context.Context, conn *jsonrpc2.Connection) (jsonrpc2.ConnectionOptions, error)

func (b binderFunc) Bind(ctx context.Context, connection *jsonrpc2.Connection) (jsonrpc2.ConnectionOptions, error) {
	return b(ctx, connection)
}
