package qilin

import (
	"encoding/json"
	"github.com/invopop/jsonschema"
	"net/url"
)

const (
	ProtocolVersion20250326 string = "2025-03-26"
	ProtocolVersion20241105 string = "2024-11-05"
	ProtocolVersion20241007 string = "2024-10-07"
	LatestProtocolVersion          = ProtocolVersion20250326
)

var SupportedProtocolVersions = map[string]bool{
	LatestProtocolVersion:   true,
	ProtocolVersion20241105: true,
	ProtocolVersion20241007: true,
}

// JSONRPCVersion represents the version of the JSON-RPC protocol used in qilin.
const JSONRPCVersion = "2.0"

const (
	// MethodInitialize Initiates connection and negotiates protocol capabilities.
	// https://modelcontextprotocol.io/specification/2024-11-05/basic/lifecycle/#initialization
	MethodInitialize string = "initialize"

	// MethodPing Verifies connection liveness between client and server.
	// https://modelcontextprotocol.io/specification/2024-11-05/basic/utilities/ping/
	MethodPing string = "ping"

	// MethodResourcesList Lists all available server resources.
	// https://modelcontextprotocol.io/specification/2024-11-05/server/resources/
	MethodResourcesList string = "resources/list"

	// MethodResourcesTemplatesList Provides URI templates for constructing resource URIs.
	// https://modelcontextprotocol.io/specification/2024-11-05/server/resources/
	MethodResourcesTemplatesList string = "resources/templates/list"

	// MethodResourcesRead retrieves content of a specific resource by URI.
	// https://modelcontextprotocol.io/specification/2024-11-05/server/resources/
	MethodResourcesRead string = "resources/read"

	// MethodResourceSubscribe Subscribes to updates for a specific resource.
	// https://modelcontextprotocol.io/specification/2025-03-26/server/resources#subscriptions
	MethodResourceSubscribe string = "resources/subscribe"

	// MethodResourceUnsubscribe Unsubscribes from updates for a specific resource.
	// https://modelcontextprotocol.io/specification/2025-03-26/server/resources#subscriptions
	MethodResourceUnsubscribe string = "resources/unsubscribe"

	// MethodPromptsList lists all available prompt templates.
	// https://modelcontextprotocol.io/specification/2024-11-05/server/prompts/
	MethodPromptsList string = "prompts/list"

	// MethodPromptsGet Retrieves a specific prompt template with filled parameters.
	// https://modelcontextprotocol.io/specification/2024-11-05/server/prompts/
	MethodPromptsGet string = "prompts/get"

	// MethodToolsList Lists all available executable tools.
	// https://modelcontextprotocol.io/specification/2024-11-05/server/tools/
	MethodToolsList string = "tools/list"

	// MethodToolsCall Invokes a specific Tool with provided parameters.
	// https://modelcontextprotocol.io/specification/2024-11-05/server/tools/
	MethodToolsCall string = "tools/call"

	//
	MethodInitializedNotification = "notifications/initialized"

	// MethodNotificationResourcesListChanged Notifies when the list of available resources changes.
	// https://modelcontextprotocol.io/specification/2025-03-26/server/resources#list-changed-notification
	MethodNotificationResourcesListChanged = "notifications/resources/list_changed"

	// MethodNotificationResourceUpdated Notifies when a specific resource is updated.
	// https://modelcontextprotocol.io/specification/2025-03-26/server/resources#subscriptions
	MethodNotificationResourceUpdated = "notifications/resources/updated"

	// NOT PLAN:
	// MethodNotificationPromptsListChanged Notifies when the list of available prompt templates changes.
	// https://modelcontextprotocol.io/specification/2025-03-26/server/prompts#list-changed-notification
	// MethodNotificationPromptsListChanged = "notifications/prompts/list_changed"

	// NOT PLAN:
	// MethodNotificationToolsListChanged Notifies when the list of available tools changes.
	// https://spec.modelcontextprotocol.io/specification/2024-11-05/server/tools/list_changed/
	// MethodNotificationToolsListChanged = "notifications/tools/list_changed"

	// MethodCompletionComplete Completes a prompt template with provided parameters.
	MethodCompletionComplete = "completion/complete"

	MethodLoggingSetLevel = "logging/setLevel"
)

// implementation describes the name and version of an MCP implementation.
type implementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// initializeRequestParams sent from the client to the server when it first connects, asking it to begin initialization.
type initializeRequestParams struct {
	// ProtocolVersion is the latest version of the Model Context Protocol that the client supports.
	//
	// The client MAY decide to support older versions as well.
	ProtocolVersion string `json:"protocolVersion"`

	Capabilities clientCapabilities `json:"capabilities"`

	ClientInfo implementation `json:"clientInfo"`
}

// clientCapabilities is a set of capabilities a client may support. Known capabilities are defined here, in this schema,
// but this is not a closed set: any client can define its own, additional capabilities.
type clientCapabilities struct {
	// Experimental is non-standard capabilities that the client supports.
	Experimental map[string]any `json:"experimental,omitzero"`

	// Sampling presents if the client supports sampling from an LLM.
	Sampling map[string]any `json:"sampling,omitzero"`

	// Roots presents if the client supports listing roots.
	Roots *RootsCapability `json:"roots,omitzero"`
}

// RootsCapability represents the client's capability to support roots features.
type RootsCapability struct {
	// ListChanged indicates whether the client supports notifications for changes to the roots list.
	ListChanged bool `json:"listChanged,omitzero"`
}

// callToolRequestParams is used by the client to invoke a Tool provided by the server.
type callToolRequestParams struct {
	// Name is the name of the Tool.
	Name string `json:"name"`

	// Arguments contains the arguments to use for the Tool.
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// NotificationsCancelledRequestParams is sent by either side to indicate that it is cancelling a previously-issued request.
type NotificationsCancelledRequestParams struct {
	// RequestID is the ID of the request to cancel.
	// This MUST correspond to the ID of a request previously issued in the same direction.
	RequestID any `json:"requestId"`
	// Reason is an optional string describing the reason for the cancellation. This MAY be logged or presented to the user.
	Reason string `json:"reason"`
}

// ServerCapabilities is a set of capabilities defined here, but this is not a closed set:
// any server can define its own, additional capabilities.
type ServerCapabilities struct {
	// Experimental contains non-standard capabilities that the server supports.
	Experimental map[string]any `json:"experimental,omitzero"`

	// Logging present if the server supports sending log messages to the client.
	Logging *LoggingCapability `json:"logging,omitzero"`

	// Completions present if the server supports argument autocompletion suggestions.
	Completions *CompletionsCapability `json:"completions,omitzero"`

	// Prompts present if the server offers any prompt templates.
	Prompts *PromptCapability `json:"prompts,omitzero"`

	// Resources present if the server offers any resources to read.
	Resources *ResourceCapability `json:"resources,omitzero"`

	// Tools present if the server offers any tools to call.
	Tools *ToolCapability `json:"tools,omitzero"`
}

// PromptCapability represents server capabilities for prompts.
type PromptCapability struct {
	// ListChanged indicates this server supports notifications for changes to the prompt list if true.
	ListChanged bool `json:"listChanged,omitzero"`
}

// ResourceCapability represents server capabilities for resources.
type ResourceCapability struct {
	// Subscribe indicates this server supports subscribing to resource updates if true.
	Subscribe bool `json:"subscribe,omitzero"`

	// ListChanged indicates this server supports notifications for changes to the resource list if true.
	ListChanged bool `json:"listChanged,omitzero"`
}

// ToolCapability represents server capabilities for tools.
type ToolCapability struct {
	// ListChanged indicates this server supports notifications for changes to the Tool list if true.
	ListChanged bool `json:"listChanged,omitzero"`
}

// LoggingCapability represents server capability for logging.
type LoggingCapability struct{}

// CompletionsCapability represents server capability for completions.
type CompletionsCapability struct{}

// initializeResult sent from the server after receiving an initialize request from the client.
type initializeResult struct {
	// ProtocolVersion is the version of the Model Context Protocol that the server wants to use.
	// This may not match the version that the client requested.
	// If the client cannot support this version, it MUST disconnect.
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      implementation     `json:"serverInfo"`

	// Instructions describe how to use the server and its features.
	//
	// This can be used by clients to improve the LLM's understanding of available tools, resources, etc.
	// It can be thought of like a "hint" to the model. For example, this information MAY be added to the system prompt.
	Instructions string `json:"instructions,omitempty"`
}

var (
	_ json.Marshaler   = (*ResourceURI)(nil)
	_ json.Unmarshaler = (*ResourceURI)(nil)
)

// ResourceURI indicates a URI to a resource or sub-resource.
type ResourceURI url.URL

func (r *ResourceURI) UnmarshalJSON(bytes []byte) error {
	var raw string
	if err := json.Unmarshal(bytes, &raw); err != nil {
		return err
	}
	uri, err := url.Parse(raw)
	if err != nil {
		return err
	}
	*r = ResourceURI(*uri)
	return nil
}

func (r ResourceURI) MarshalJSON() ([]byte, error) {
	uri := url.URL(r)
	return json.Marshal(uri.String())
}

type ResourceContent interface {
	// GetURI returns the URI of the resource.
	GetURI() url.URL
	// GetMimeType returns the MIME type of the resource.
	GetMimeType() string
	json.Marshaler
}

// resourceContentBase the contents of a specific resource or sub-resource.
type resourceContentBase struct {
	// uri of this resource.
	uri url.URL

	// MIME type of this resource, if known.
	mimeType string
}

// compatibility check
var _ ResourceContent = (*textResourceContent)(nil)

// textResourceContent is a resource content that contains text.
type textResourceContent struct {
	resourceContentBase

	// text of the item. This must only be set if the item can actually be represented as text (not binary data).
	text    string
	marshal JSONMarshalFunc
}

func (t textResourceContent) MarshalJSON() ([]byte, error) {
	return t.marshal(struct {
		URI      string `json:"uri"`
		MimeType string `json:"mimeType,omitzero"`
		Text     string `json:"text,omitzero"`
	}{
		URI:      t.uri.String(),
		MimeType: t.mimeType,
		Text:     t.text,
	})
}

func (t textResourceContent) GetURI() url.URL {
	return t.uri
}

func (t textResourceContent) GetMimeType() string {
	return t.mimeType
}

// compatibility check
var _ ResourceContent = (*binaryResourceContent)(nil)

// binaryResourceContent is a resource content that contains binary data.
type binaryResourceContent struct {
	resourceContentBase

	// blob is a base64-encoded string representing the binary data of the item.
	blob    string
	marshal JSONMarshalFunc
}

func (b binaryResourceContent) MarshalJSON() ([]byte, error) {
	return b.marshal(struct {
		URI      string `json:"uri"`
		MimeType string `json:"mimeType,omitzero"`
		Blob     string `json:"blob"`
	}{
		URI:      b.uri.String(),
		MimeType: b.mimeType,
		Blob:     b.blob,
	})
}

func (b binaryResourceContent) GetURI() url.URL {
	return b.uri
}

func (b binaryResourceContent) GetMimeType() string {
	return b.mimeType
}

// readResourceResult is the server's response to a resources/read request from the client.
type readResourceResult struct {
	// Contents is the content of the resource.
	Contents []ResourceContent `json:"contents"`
}

// Resource that the server is capable of reading.
type Resource struct {
	// URI of this resource.
	URI *ResourceURI `json:"uri"`

	// Name of the resource that is human-readable.
	//
	// This can be used by clients to populate UI elements.
	Name string `json:"name"`

	// Description of what this resource represents.
	//
	// This can be used by clients to improve the LLM's understanding of available resources.
	// It can be thought of like a "hint" to the model.
	Description string `json:"description,omitzero"`

	// MimeType of this resource, if known.
	MimeType string `json:"mimeType,omitzero"`

	// handler handles reading the resource.
	handler ResourceHandlerFunc `json:"-"`

	// resourceChangeCtx can be used to subscribe to changes to this resource.
	resourceChangeCtx ResourceChangeContext `json:"-"`
}

// resourceTemplate a template description for resources available on the server.
type resourceTemplate struct {
	// URITemplate can be used to construct resource URIs. according to RFC 6570
	URITemplate *ResourceURI `json:"uriTemplate"`

	// Name for the type of resource this template refers to.
	//
	// This can be used by clients to populate UI elements.
	Name string `json:"name"`

	// Description of what this template is for.
	//
	// This can be used by clients to improve the LLM's understanding of available resources.
	// It can be thought of like a "hint" to the model.
	Description string `json:"description,omitzero"`

	// MimeType for all resources that match this template.
	// This should only be included if all resources matching this template have the same type.
	MimeType string `json:"mimeType,omitzero"`
}

// listResourcesResult is the server's response to a request for a list of resources.
type listResourcesResult struct {
	// Resources is a list of resources available on the server.
	Resources []Resource `json:"resources"`
}

// listResourceTemplatesResult is the server's response to a request for a list of resource templates.
type listResourceTemplatesResult struct {
	// ResourceTemplates is a list of resource templates available on the server.
	ResourceTemplates []resourceTemplate `json:"resourceTemplates"`
}

// readResourceRequestParams sent from the client to the server to read a specific resource URI.
type readResourceRequestParams struct {
	// URI of the resource to read. The URI can use any protocol; it is up to the server how to interpret it.
	URI *ResourceURI `json:"uri"`
}

// Tool defines a Tool that the client can call.
type Tool struct {
	// Name of the Tool.
	Name string `json:"name"`

	// Description of the Tool that is human-readable.
	Description string `json:"description,omitzero"`

	// InputSchema defines the arguments that the Tool accepts in JSON Schema format.
	InputSchema *jsonschema.Schema `json:"inputSchema"`

	// Annotations hint to the client about the Tool's behavior.
	Annotations *ToolAnnotations `json:"annotations,omitzero"`

	// handler handles invoke the Tool with the provided arguments.
	handler ToolHandlerFunc `json:"-"`
}

// ToolAnnotations represents additional properties describing a Tool to clients.
//
// NOTE: all properties in ToolAnnotations are **hints**.
// They are not guaranteed to provide a faithful description of
// Tool behavior (including descriptive properties like `title`).
type ToolAnnotations struct {
	// Title is a human-readable title for the Tool.
	Title string `json:"title,omitzero"`

	// ReadOnlyHint indicates the Tool does not modify its environment if true
	//
	// Default: false
	ReadOnlyHint bool `json:"readOnlyHint,omitzero"`

	// DestructiveHint indicates whether the Tool may perform destructive updates to its environment if true.
	// If false, the Tool performs only additive updates.
	//
	// (This property is meaningful only when `readOnlyHint == false`)
	//
	// Default: true
	DestructiveHint bool `json:"destructiveHint,omitzero"`

	// IdempotentHint indicates that calling the Tool repeatedly with the same arguments
	// will have no additional effect on the its environment if true.
	//
	// (This property is meaningful only when `readOnlyHint == false`)
	//
	// Default: false
	IdempotentHint bool `json:"idempotentHint,omitzero"`

	// OpenWorldHint indicates this Tool may interact with an "open world" of external entities if true.
	// If false, the Tool's domain of interaction is closed.
	// For example, the world of a web search Tool is open, whereas that
	// of a memory Tool is not.
	//
	// Default: true
	OpenWorldHint bool `json:"openWorldHint,omitzero"`
}

type listToolsResponse struct {
	NextCursor string `json:"nextCursor,omitzero"`
	Tools      []Tool `json:"tools"`
}

type CallToolContent interface {
	GetType() string
}

// compatibility check
var _ CallToolContent = (*textCallToolContent)(nil)

type textCallToolContent struct {
	Text        string
	Annotations *ToolAnnotations
	marshal     JSONMarshalFunc
}

func (t *textCallToolContent) MarshalJSON() ([]byte, error) {
	return t.marshal(struct {
		Type        string           `json:"type"`
		Text        string           `json:"text"`
		Annotations *ToolAnnotations `json:"annotations,omitzero"`
	}{
		Type:        t.GetType(),
		Text:        t.Text,
		Annotations: t.Annotations,
	})
}

func (t *textCallToolContent) GetType() string {
	return "text"
}

// compatibility check
var _ CallToolContent = (*imageCallToolContent)(nil)

type imageCallToolContent struct {
	Data     string
	MimeType string
	marshal  JSONMarshalFunc
}

func (i *imageCallToolContent) MarshalJSON() ([]byte, error) {
	return i.marshal(struct {
		Type     string `json:"type"`
		Data     string `json:"data"`
		MimeType string `json:"mimeType,omitzero"`
	}{
		Type:     i.GetType(),
		Data:     i.Data,
		MimeType: i.MimeType,
	})
}

func (i *imageCallToolContent) GetType() string {
	return "image"
}

// compatibility check
var _ CallToolContent = (*audioCallToolContent)(nil)

type audioCallToolContent struct {
	Data     string
	MimeType string
	marshal  JSONMarshalFunc
}

func (a *audioCallToolContent) MarshalJSON() ([]byte, error) {
	return a.marshal(struct {
		Type     string `json:"type"`
		Data     string `json:"data"`
		MimeType string `json:"mimeType,omitzero"`
	}{
		Type:     a.GetType(),
		Data:     a.Data,
		MimeType: a.MimeType,
	})
}

func (a *audioCallToolContent) GetType() string {
	return "audio"
}

// compatibility check
var _ CallToolContent = (*embedResourceCallToolContent)(nil)

type embedResourceCallToolContent struct {
	Resource ResourceContent
	marshal  JSONMarshalFunc
}

func (e *embedResourceCallToolContent) MarshalJSON() ([]byte, error) {
	return e.marshal(struct {
		Type     string          `json:"type"`
		Resource ResourceContent `json:"resource"`
	}{
		Type:     e.GetType(),
		Resource: e.Resource,
	})
}

func (e *embedResourceCallToolContent) GetType() string {
	return "resource"
}

// subscribeResourcesRequestParams sent from the client to request resources/updated notifications from the server whenever a particular resource changes.
type subscribeResourcesRequestParams struct {
	URI *ResourceURI `json:"uri"`
}

// unsubscribeResourcesRequestParams from the client to request cancellation of resources/updated notifications from the server.
//
// This should follow a previous resources/subscribe request
type unsubscribeResourcesRequestParams struct {
	URI *ResourceURI `json:"uri"`
}

// resourceUpdatedNotificationParams sent from the server to the client, informing it that a resource has changed and may need to be read again.
// This should only be sent if the client previously sent a resources/subscribe request.
type resourceUpdatedNotificationParam struct {
	// URI of the resource that changed.
	URI *ResourceURI `json:"uri"`
}
