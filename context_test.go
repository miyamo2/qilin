package qilin

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
	"weak"

	"golang.org/x/exp/jsonrpc2"
)

func Test_context_Get(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		c := &_context{}
		key := "key"
		value := "value"
		c.Set(key, value)

		if got := c.Get(key); got != value {
			t.Fatalf("expected %v, got %v", value, got)
		}
	})
	t.Run("unset", func(t *testing.T) {
		c := &_context{}
		key := "unsetKey"

		if got := c.Get(key); got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})
}

func Test_context_Set(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		c := &_context{}
		key := "key"
		value := "value"
		c.Set(key, value)

		if got, _ := c.store.Load(key); got != value {
			t.Fatalf("expected %v, got %v", value, got)
		}
	})
	t.Run("nil value", func(t *testing.T) {
		c := &_context{}
		key := "key"
		var value interface{}
		c.Set(key, value)

		if got, _ := c.store.Load(key); got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})
}

func Test_context_Context(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		c := &_context{
			ctx: t.Context(),
		}
		if got := c.Context(); got != c.ctx {
			t.Fatalf("expected %v, got %v", c.ctx, got)
		}
	})
	t.Run("unset", func(t *testing.T) {
		c := &_context{}
		if got := c.Context(); got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})
}

func Test_context_SetContext(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		c := &_context{}
		ctx := t.Context()
		c.SetContext(ctx)

		if got := c.Context(); got != ctx {
			t.Fatalf("expected %v, got %v", ctx, got)
		}
	})
	t.Run("nil context", func(t *testing.T) {
		c := &_context{}
		//nolint:staticcheck
		//lint:ignore SA1012 just for test
		c.SetContext(nil)

		if got := c.Context(); got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})
}

func Test_context_reset(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		c := &_context{
			ctx: t.Context(),
			jsonrpcRequest: &jsonrpc2.Request{
				ID:     jsonrpc2.StringID("1"),
				Method: "method",
				Params: json.RawMessage(`{"x": 1.5, "y": 2.5}`),
			},
			jsonMarshalFunc:   json.Marshal,
			jsonUnmarshalFunc: json.Unmarshal,
		}
		key := "key"
		value := "value"
		c.Set(key, value)

		c.reset()
		if c.jsonrpcRequest != nil {
			t.Fatalf("expected nil jsonrpcRequest, got %v", c.jsonrpcRequest)
		}
		if c.jsonMarshalFunc == nil {
			t.Fatalf("expected jsonMarshalFunc  to be not null, but null.")
		}
		if c.jsonUnmarshalFunc == nil {
			t.Fatalf("expected jsonUnmarshalFunc to be not null, but null.")
		}
		c.store.Range(func(k, v interface{}) bool {
			t.Fatalf("unexpected key=%v value=%v %v", k, v, v)
			return true
		})
		if got := c.ctx; got != nil {
			t.Fatalf("expected nil context, got %v", got)
		}

	})
}

func Test_context_JSONRPCRequest(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		req := jsonrpc2.Request{
			ID:     jsonrpc2.StringID("1"),
			Method: "method",
			Params: json.RawMessage(`{"x": 1.5, "y": 2.5}`),
		}
		c := &_context{
			jsonrpcRequest: &req,
		}
		if !reflect.DeepEqual(c.JSONRPCRequest(), req) {
			t.Fatalf("expected %v, got %v", req, c.JSONRPCRequest())
		}
	})
	t.Run("unset", func(t *testing.T) {
		empty := jsonrpc2.Request{}
		c := &_context{
			jsonrpcRequest: nil,
		}
		if !reflect.DeepEqual(c.JSONRPCRequest(), empty) {
			t.Fatalf("expected %v, got %v", empty, c.JSONRPCRequest())
		}
	})
}

func TestToolContext_ToolName(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		c := newToolContext(nil, nil, nil)
		c.toolName = "test"
		if got := c.ToolName(); got != "test" {
			t.Fatalf("expected 'test', got %v", got)
		}
	})
	t.Run("unset", func(t *testing.T) {
		c := newToolContext(nil, nil, nil)
		if got := c.ToolName(); got != "" {
			t.Fatalf("expected empty string, got %v", got)
		}
	})
}

func TestToolContext_Arguments(t *testing.T) {
	args := json.RawMessage(`{"x": 1.5, "y": 2.5}`)
	t.Run("happy path", func(t *testing.T) {
		c := newToolContext(nil, nil, nil)
		c.args = args
		if got := c.Arguments(); !reflect.DeepEqual(got, args) {
			t.Fatalf("expected %v, got %v", args, got)
		}
	})
}

func TestToolContext_Bind(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		type Req struct {
			X float64 `json:"x" jsonschema:"title=X"`
			Y float64 `json:"y" jsonschema:"title=Y"`
		}
		c := newToolContext(json.Unmarshal, nil, nil)
		c.args = json.RawMessage(`{"x": 1.5, "y": 2.5}`)
		var req Req
		err := c.Bind(&req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if req.X != 1.5 || req.Y != 2.5 {
			t.Fatalf("expected x=1.5, y=2.5, got x=%v, y=%v", req.X, req.Y)
		}
	})
	t.Run("empty args", func(t *testing.T) {
		type Req struct {
			X float64 `json:"x" jsonschema:"title=X"`
			Y float64 `json:"y" jsonschema:"title=Y"`
		}
		c := newToolContext(json.Unmarshal, nil, nil)
		var req Req
		err := c.Bind(&req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if req.X != 0 || req.Y != 0 {
			t.Fatalf("expected x=0, y=0, got x=%v, y=%v", req.X, req.Y)
		}
	})
}

func TestToolContext_String(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		var dest CallToolContent
		c := newToolContext(nil, nil, nil)
		c.dest = &dest
		c.jsonrpcRequest = &jsonrpc2.Request{
			ID:     jsonrpc2.StringID("1"),
			Method: "method",
			Params: json.RawMessage(`{"x": 1.5, "y": 2.5}`),
		}
		if err := c.String("hello"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v, ok := (dest).(*textCallToolContent)
		if !ok {
			t.Fatalf("expected *textCallToolContent, got %T", dest)
		}
		if v.Text != "hello" {
			t.Fatalf("expected 'hello', got %v", v.Text)
		}
	})
}

func TestToolContext_JSON(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		type Res struct {
			Result float64 `json:"result"`
		}
		var dest CallToolContent
		c := newToolContext(nil, json.Marshal, nil)
		c.dest = &dest
		res := Res{Result: 3.0}
		if err := c.JSON(res); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v, ok := (dest).(*textCallToolContent)
		if !ok {
			t.Fatalf("expected *textCallToolContent, got %T", dest)
		}

		expect := `{"result":3}`
		if strings.ReplaceAll(v.Text, " ", "") != expect {
			t.Fatalf("expected `%s`, got %v", expect, v.Text)
		}
	})
	t.Run("unmarshal error", func(t *testing.T) {
		type Res struct {
			Result float64 `json:"result"`
		}
		errTest := fmt.Errorf("test error")
		var dest CallToolContent
		c := newToolContext(nil, func(v interface{}) ([]byte, error) {
			return nil, errTest
		}, nil)
		c.dest = &dest
		res := Res{Result: 3.0}
		if err := c.JSON(res); !errors.Is(err, errTest) {
			t.Fatalf("expected error %v, got %v", errTest, err)
		}
	})
}

func TestToolContext_Audio(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		var dest CallToolContent
		c := newToolContext(nil, nil, func(data []byte) string {
			return "test"
		})
		c.dest = &dest
		var audio []byte
		mime := "audio/wav"
		if err := c.Audio(audio, mime); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v, ok := (dest).(*audioCallToolContent)
		if !ok {
			t.Fatalf("expected *audioCallToolContent, got %T", dest)
		}
		if v.Data != "test" {
			t.Fatalf("expected 'test', got %v", v.Data)
		}
		if v.MimeType != mime {
			t.Fatalf("expected '%s', got %v", mime, v.MimeType)
		}
	})
}

func TestToolContext_Image(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		var dest CallToolContent
		c := newToolContext(nil, nil, func(data []byte) string {
			return "test"
		})
		c.dest = &dest
		var image []byte
		mime := "image/png"
		if err := c.Image(image, mime); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v, ok := (dest).(*imageCallToolContent)
		if !ok {
			t.Fatalf("expected *imageCallToolContent, got %T", dest)
		}
		if v.Data != "test" {
			t.Fatalf("expected 'test', got %v", v.Data)
		}
		if v.MimeType != mime {
			t.Fatalf("expected '%s', got %v", mime, v.MimeType)
		}
	})
}

func TestToolContext_StringResource(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		var dest CallToolContent
		c := newToolContext(nil, nil, nil)
		c.dest = &dest
		uri := MustURL(t, "example://example.com")
		if err := c.StringResource(uri, "test", ""); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v, ok := (dest).(*embedResourceCallToolContent)
		if !ok {
			t.Fatalf("expected *embedResourceCallToolContent, got %T", dest)
		}
		resource, ok := v.Resource.(*textResourceContent)
		if !ok {
			t.Fatalf("expected *textResourceContent, got %T", v.Resource)
		}
		if !reflect.DeepEqual(resource.uri.Value(), uri) {
			t.Fatalf("expected '%v', got %v", uri, resource.uri.Value())
		}
		if resource.mimeType != "text/plain" {
			t.Fatalf("expected 'text/plain', got %v", resource.mimeType)
		}
		if resource.text != "test" {
			t.Fatalf("expected 'test', got %v", resource.text)
		}
	})
	t.Run("with mime type", func(t *testing.T) {
		var dest CallToolContent
		c := newToolContext(nil, nil, nil)
		c.dest = &dest
		uri := MustURL(t, "example://example.com")
		if err := c.StringResource(uri, "test", "text/html"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v, ok := (dest).(*embedResourceCallToolContent)
		if !ok {
			t.Fatalf("expected *embedResourceCallToolContent, got %T", dest)
		}
		resource, ok := v.Resource.(*textResourceContent)
		if !ok {
			t.Fatalf("expected *textResourceContent, got %T", v.Resource)
		}
		if resource.mimeType != "text/html" {
			t.Fatalf("expected 'text/html', got %v", resource.mimeType)
		}
	})
}

func TestToolContext_JSONResource(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		var dest CallToolContent
		c := newToolContext(nil, json.Marshal, nil)
		c.dest = &dest
		uri := MustURL(t, "example://example.com")
		if err := c.JSONResource(uri, map[string]string{"key": "value"}, ""); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v, ok := (dest).(*embedResourceCallToolContent)
		if !ok {
			t.Fatalf("expected *embedResourceCallToolContent, got %T", dest)
		}
		resource, ok := v.Resource.(*textResourceContent)
		if !ok {
			t.Fatalf("expected *textResourceContent, got %T", v.Resource)
		}
		if resource.mimeType != "application/json" {
			t.Fatalf("expected 'application/json', got %v", resource.mimeType)
		}
		if resource.text != `{"key":"value"}` {
			t.Fatalf("expected '{\"key\":\"value\"}', got %v", resource.text)
		}
	})
	t.Run("with mime type", func(t *testing.T) {
		var dest CallToolContent
		c := newToolContext(nil, json.Marshal, nil)
		c.dest = &dest
		uri := MustURL(t, "example://example.com")
		mime := "application/ld+json"
		if err := c.JSONResource(uri, map[string]string{"key": "value"}, mime); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v, ok := (dest).(*embedResourceCallToolContent)
		if !ok {
			t.Fatalf("expected *embedResourceCallToolContent, got %T", dest)
		}
		resource, ok := v.Resource.(*textResourceContent)
		if !ok {
			t.Fatalf("expected *textResourceContent, got %T", v.Resource)
		}
		if resource.mimeType != mime {
			t.Fatalf("expected '%s', got %v", mime, resource.mimeType)
		}
	})
	t.Run("unmarshal error", func(t *testing.T) {
		var dest CallToolContent
		errTest := fmt.Errorf("test error")
		c := newToolContext(nil, func(v interface{}) ([]byte, error) {
			return nil, errTest
		}, nil)
		c.dest = &dest
		uri := MustURL(t, "example://example.com")
		if err := c.JSONResource(uri, map[string]string{"key": "value"}, ""); !errors.Is(
			err,
			errTest,
		) {
			t.Fatalf("expected error %v, got %v", errTest, err)
		}
	})
}

func TestToolContext_BinaryResource(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		var dest CallToolContent
		c := newToolContext(nil, nil, func(data []byte) string {
			return "test"
		})
		c.dest = &dest
		var binary []byte
		uri := MustURL(t, "example://example.com")
		mime := "application/octet-stream"
		if err := c.BinaryResource(uri, binary, ""); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v, ok := (dest).(*embedResourceCallToolContent)
		if !ok {
			t.Fatalf("expected *embedResourceCallToolContent, got %T", dest)
		}
		resource, ok := v.Resource.(*binaryResourceContent)
		if !ok {
			t.Fatalf("expected *binaryResourceContent, got %T", v.Resource)
		}
		if resource.mimeType != mime {
			t.Fatalf("expected '%s', got %v", mime, resource.mimeType)
		}
		if resource.blob != "test" {
			t.Fatalf("expected 'test', got %v", resource.blob)
		}
	})
	t.Run("with mime type", func(t *testing.T) {
		var dest CallToolContent
		c := newToolContext(nil, nil, func(data []byte) string {
			return "test"
		})
		c.dest = &dest
		var binary []byte
		uri := MustURL(t, "example://example.com")
		mime := "application/pdf"
		if err := c.BinaryResource(uri, binary, mime); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v, ok := (dest).(*embedResourceCallToolContent)
		if !ok {
			t.Fatalf("expected *embedResourceCallToolContent, got %T", dest)
		}
		resource, ok := v.Resource.(*binaryResourceContent)
		if !ok {
			t.Fatalf("expected *binaryResourceContent, got %T", v.Resource)
		}
		if resource.mimeType != mime {
			t.Fatalf("expected '%s', got %v", mime, resource.mimeType)
		}
	})
}

func TestToolContext_reset(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		var dest CallToolContent
		c := newToolContext(json.Unmarshal, json.Marshal, base64.StdEncoding.EncodeToString)
		c.toolName = "test"
		c.args = json.RawMessage(`{"x": 1.5, "y": 2.5}`)
		c.dest = &dest
		c.annotation = &ToolAnnotations{}
		c.jsonrpcRequest = &jsonrpc2.Request{}
		c.store.Store("key", "value")
		c.reset()
		if c.toolName != "" {
			t.Fatalf("expected empty string, got %v", c.toolName)
		}
		if c.args != nil {
			t.Fatalf("expected nil, got %v", c.args)
		}
		if c.dest != nil {
			t.Fatalf("expected nil, got %v", c.dest)
		}
		if c.annotation != nil {
			t.Fatalf("expected nil, got %v", c.annotation)
		}
		if c.jsonrpcRequest != nil {
			t.Fatalf("expected nil, got %v", c.jsonrpcRequest)
		}
		c.store.Range(func(k, v interface{}) bool {
			t.Fatalf("unexpected key=%v value=%v %v", k, v, v)
			return true
		})
		if c.jsonMarshalFunc == nil {
			t.Fatalf("expected jsonMarshalFunc to be not null, but null.")
		}
		if c.jsonUnmarshalFunc == nil {
			t.Fatalf("expected jsonUnmarshalFunc to be not null, but null.")
		}
		if c.base64StringFunc == nil {
			t.Fatalf("expected base64StringFunc to be not null, but null.")
		}
	})
}

func TestResourceContext_ResourceURI(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		uri := MustURL(t, "example://example.com")
		c := newResourceContext(nil, nil, nil)
		c.uri = weak.Make(uri)
		if got := c.ResourceURI(); !reflect.DeepEqual(got, uri) {
			t.Fatalf("expected %v, got %v", uri, got)
		}
	})
	t.Run("unset", func(t *testing.T) {
		c := newResourceContext(nil, nil, nil)
		if got := c.ResourceURI(); got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})
}

func TestResourceContext_MimeType(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		c := newResourceContext(nil, nil, nil)
		c.mimeType = "application/json"
		if got := c.MimeType(); got != "application/json" {
			t.Fatalf("expected 'application/json', got %v", got)
		}
	})
	t.Run("unset", func(t *testing.T) {
		c := newResourceContext(nil, nil, nil)
		if got := c.MimeType(); got != "" {
			t.Fatalf("expected empty string, got %v", got)
		}
	})
}

func TestResourceContext_Param(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		c := newResourceContext(nil, nil, nil)
		c.pathParams = map[string]string{"id": "123"}
		if got := c.Param("id"); got != "123" {
			t.Fatalf("expected 123, got %v", got)
		}
	})
	t.Run("unset key", func(t *testing.T) {
		c := newResourceContext(nil, nil, nil)
		c.pathParams = map[string]string{"id": "123"}
		if got := c.Param("unsetKey"); got != "" {
			t.Fatalf("expected empty string, got %v", got)
		}
	})
}

func TestResourceContext_String(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		c := newResourceContext(nil, nil, nil)
		c.dest = &readResourceResult{}
		if err := c.String("hello"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(c.dest.Contents) != 1 {
			t.Fatalf("expected 1 content, got %d", len(c.dest.Contents))
		}
		v, ok := (c.dest.Contents[0]).(textResourceContent)
		if !ok {
			t.Fatalf("expected textResourceContent, got %T", c.dest.Contents[0])
		}
		if v.text != "hello" {
			t.Fatalf("expected 'hello', got %v", v.text)
		}
	})
	t.Run("with mime type", func(t *testing.T) {
		c := newResourceContext(nil, nil, nil)
		c.dest = &readResourceResult{}
		mime := "text/csv"
		c.mimeType = mime
		if err := c.String("hello"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(c.dest.Contents) != 1 {
			t.Fatalf("expected 1 content, got %d", len(c.dest.Contents))
		}
		v, ok := (c.dest.Contents[0]).(textResourceContent)
		if !ok {
			t.Fatalf("expected textResourceContent, got %T", c.dest.Contents[0])
		}
		if v.mimeType != mime {
			t.Fatalf("expected '%s', got %v", mime, v.mimeType)
		}
	})
}

func TestResourceContext_JSON(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		type Res struct {
			Result float64 `json:"result"`
		}
		c := newResourceContext(nil, json.Marshal, nil)
		c.dest = &readResourceResult{}
		res := Res{Result: 3.0}
		if err := c.JSON(res); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(c.dest.Contents) != 1 {
			t.Fatalf("expected 1 content, got %d", len(c.dest.Contents))
		}
		v, ok := (c.dest.Contents[0]).(textResourceContent)
		if !ok {
			t.Fatalf("expected textResourceContent, got %T", c.dest.Contents[0])
		}

		expect := `{"result":3}`
		if strings.ReplaceAll(v.text, " ", "") != expect {
			t.Fatalf("expected `%s`, got %v", expect, v.text)
		}
	})
	t.Run("with mime type", func(t *testing.T) {
		type Res struct {
			Result float64 `json:"result"`
		}
		c := newResourceContext(nil, json.Marshal, nil)
		c.dest = &readResourceResult{}
		mime := "application/ld+json"
		c.mimeType = mime
		res := Res{Result: 3.0}
		if err := c.JSON(res); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(c.dest.Contents) != 1 {
			t.Fatalf("expected 1 content, got %d", len(c.dest.Contents))
		}
		v, ok := (c.dest.Contents[0]).(textResourceContent)
		if !ok {
			t.Fatalf("expected textResourceContent, got %T", c.dest.Contents[0])
		}
		if v.mimeType != mime {
			t.Fatalf("expected '%s', got %v", mime, v.mimeType)
		}
	})
	t.Run("unmarshal error", func(t *testing.T) {
		type Res struct {
			Result float64 `json:"result"`
		}
		errTest := fmt.Errorf("test error")
		c := newResourceContext(nil, func(v interface{}) ([]byte, error) {
			return nil, errTest
		}, nil)
		c.dest = &readResourceResult{}
		res := Res{Result: 3.0}
		if err := c.JSON(res); !errors.Is(err, errTest) {
			t.Fatalf("expected error %v, got %v", errTest, err)
		}
	})
}

func TestResourceContext_Blob(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		c := newResourceContext(nil, nil, func(data []byte) string {
			return "test"
		})
		c.dest = &readResourceResult{}
		var blob []byte
		mime := "application/octet-stream"
		if err := c.Blob(blob, ""); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v, ok := (c.dest.Contents[0]).(binaryResourceContent)
		if !ok {
			t.Fatalf("expected binaryResourceContent, got %T", c.dest.Contents[0])
		}
		if v.mimeType != mime {
			t.Fatalf("expected '%s', got %v", mime, v.mimeType)
		}
		if v.blob != "test" {
			t.Fatalf("expected 'test', got %v", v.blob)
		}
	})
	t.Run("mime type set in field", func(t *testing.T) {
		c := newResourceContext(nil, nil, func(data []byte) string {
			return "test"
		})
		c.dest = &readResourceResult{}
		c.mimeType = "application/pdf"
		var blob []byte
		mime := "application/pdf"
		if err := c.Blob(blob, ""); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v, ok := (c.dest.Contents[0]).(binaryResourceContent)
		if !ok {
			t.Fatalf("expected binaryResourceContent, got %T", c.dest.Contents[0])
		}
		if v.mimeType != mime {
			t.Fatalf("expected '%s', got %v", mime, v.mimeType)
		}
	})
	t.Run("", func(t *testing.T) {})
}

func TestResourceContext_reset(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		c := newResourceContext(json.Unmarshal, json.Marshal, base64.StdEncoding.EncodeToString)
		c.uri = weak.Make(MustURL(t, "example://example.com"))
		c.jsonrpcRequest = &jsonrpc2.Request{
			ID:     jsonrpc2.StringID("1"),
			Method: "method",
			Params: json.RawMessage(`{"x": 1.5, "y": 2.5}`),
		}
		c.store.Store("key", "value")
		c.mimeType = "application/json"
		c.pathParams = map[string]string{"id": "123"}
		c.dest = &readResourceResult{}
		c.reset()
		if c.mimeType != "" {
			t.Fatalf("expected empty string, got %v", c.mimeType)
		}
		if len(c.pathParams) != 0 {
			t.Fatalf("expected nil, got %v", c.pathParams)
		}
		if c.dest != nil {
			t.Fatalf("expected nil, got %v", c.dest)
		}
		if c.jsonrpcRequest != nil {
			t.Fatalf("expected nil, got %v", c.jsonrpcRequest)
		}
		c.store.Range(func(k, v interface{}) bool {
			t.Fatalf("unexpected key=%v value=%v %v", k, v, v)
			return true
		})
	})
}

func TestResourceListContext_Resources(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		resources := map[string]Resource{
			"example://example.com/1": {
				URI:         (*ResourceURI)(MustURL(t, "example://example.com/1")),
				MimeType:    "application/json",
				Name:        "example",
				Description: "example description",
			},
		}
		c := newResourceListContext(nil, nil)
		c.resources = resources
		got := c.Resources()
		if !reflect.DeepEqual(got, resources) {
			t.Fatalf("expected %v, got %v", resources, got)
		}
	})
}

func TestResourceListContext_SetResource(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		c := newResourceListContext(nil, nil)
		uri := MustURL(t, "example://example.com/1")
		dest := map[string]Resource{}
		c.dest = &dest
		resource := Resource{
			URI:         (*ResourceURI)(uri),
			MimeType:    "application/json",
			Name:        "example",
			Description: "example description",
		}
		c.SetResource(uri.String(), resource)
		if got, ok := dest[uri.String()]; !ok || !reflect.DeepEqual(got, resource) {
			t.Fatalf("expected %v, got %v", resource, got)
		}
	})
}

func TestResourceListContext_reset(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		c := newResourceListContext(json.Unmarshal, json.Marshal)
		c.jsonrpcRequest = &jsonrpc2.Request{
			ID:     jsonrpc2.StringID("1"),
			Method: "method",
			Params: json.RawMessage(`{"x": 1.5, "y": 2.5}`),
		}
		c.store.Store("key", "value")
		c.resources = map[string]Resource{}
		c.reset()
		if c.jsonrpcRequest != nil {
			t.Fatalf("expected nil, got %v", c.jsonrpcRequest)
		}
		c.store.Range(func(k, v interface{}) bool {
			t.Fatalf("unexpected key=%v value=%v %v", k, v, v)
			return true
		})
	})
}

func TestResourceChangeSubscriber_ID(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		s := resourceChangeSubscriber{
			id: "test-id",
		}
		if got := s.ID(); got != "test-id" {
			t.Fatalf("expected 'test-id', got %v", got)
		}
	})
}

func TestResourceChangeSubscriber_LastReceived(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		time := MustTime(t, "2023-10-01T00:00:00Z")
		s := resourceChangeSubscriber{
			lastReceived: time,
		}
		if got := s.LastReceived(); !got.Equal(time) {
			t.Fatalf("expected %v, got %v", time, got)
		}
	})
	t.Run("unset", func(t *testing.T) {
		time := MustTime(t, "0001-01-01T00:00:00Z")
		s := resourceChangeSubscriber{}
		if got := s.LastReceived(); !got.Equal(time) {
			t.Fatalf("expected %v, got %v", time, got)
		}
	})
}

func TestResourceChangeSubscriber_SubscribedURI(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		uri := MustURL(t, "example://example.com")
		s := resourceChangeSubscriber{
			subscribedURI: uri,
		}
		if got := s.SubscribedURI(); !reflect.DeepEqual(got, uri) {
			t.Fatalf("expected %v, got %v", uri, got)
		}
	})
	t.Run("unset", func(t *testing.T) {
		s := resourceChangeSubscriber{}
		if got := s.SubscribedURI(); got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})
}

func TestResourceChangeSubscriber_Publish(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		ch := make(chan *url.URL, 1)
		defer close(ch)
		uri := MustURL(t, "example://example.com")
		lastReceived := MustTime(t, "2023-10-01T00:00:00Z")
		s := resourceChangeSubscriber{
			subscribedURI: uri,
			lastReceived:  lastReceived,
			ch:            ch,
			nowFunc: func() time.Time {
				return lastReceived.Add(1)
			},
		}

		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()

		s.Publish(MustURL(t, "example://example.com"))
		select {
		case <-ctx.Done():
			t.Fatalf("expected to receive message, got timeout")
		case msg := <-ch:
			if msg == nil {
				t.Fatalf("expected non-nil message, got nil")
			}
			if msg.String() != uri.String() {
				t.Fatalf("expected 'example://example.com', got %v", msg)
			}
		}
	})
}

func TestResourceChangeSubscriber_reset(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		ch := make(chan *url.URL, 1)
		uri := MustURL(t, "example://example.com")
		lastReceived := MustTime(t, "2023-10-01T00:00:00Z")
		s := resourceChangeSubscriber{
			subscribedURI: uri,
			lastReceived:  lastReceived,
			ch:            ch,
			nowFunc: func() time.Time {
				return lastReceived.Add(1)
			},
		}
		s.reset()
		if s.subscribedURI != nil {
			t.Fatalf("expected nil, got %v", s.subscribedURI)
		}
		if s.lastReceived != (time.Time{}) {
			t.Fatalf("expected zero time, got %v", s.lastReceived)
		}
		if s.ch != nil {
			t.Fatalf("expected nil, got %v", s.ch)
		}
		if s.nowFunc == nil {
			t.Fatalf("expected non-nil, got nil")
		}
	})
}

func TestResourceChangeContext_Context(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		c := resourceChangeContext{
			ctx: t.Context(),
		}
		if got := c.Context(); got != t.Context() {
			t.Fatalf("expected %v, got %v", t.Context(), got)
		}
	})
}

func TestResourceChangeContext_Publish(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		ch := make(chan *url.URL, 1)
		defer close(ch)
		uri := MustURL(t, "example://example.com")
		now := MustTime(t, "2023-10-01T00:00:00Z")

		c := resourceChangeContext{
			ctx: t.Context(),
			subscriber: map[string]ResourceChangeSubscriber{
				"1": &resourceChangeSubscriber{
					id:            "1",
					subscribedURI: uri,
					lastReceived:  now,
					ch:            ch,
					nowFunc: func() time.Time {
						return now.Add(1)
					},
				},
			},
		}
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()

		c.Publish(uri, now.Add(1))
		select {
		case <-ctx.Done():
			t.Fatalf("expected to receive message, got timeout")
		case msg := <-ch:
			if msg == nil {
				t.Fatalf("expected non-nil message, got nil")
			}
			if msg.String() != uri.String() {
				t.Fatalf("expected 'example://example.com', got %v", msg)
			}
		}
	})
	t.Run("with path params", func(t *testing.T) {
		ch := make(chan *url.URL, 1)
		defer close(ch)
		subscribeURI := MustURL(t, "example://example.com/{id}")
		now := MustTime(t, "2023-10-01T00:00:00Z")

		c := resourceChangeContext{
			ctx: t.Context(),
			subscriber: map[string]ResourceChangeSubscriber{
				"1": &resourceChangeSubscriber{
					id:            "1",
					subscribedURI: subscribeURI,
					lastReceived:  now,
					ch:            ch,
					nowFunc: func() time.Time {
						return now.Add(1)
					},
				},
			},
		}
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()

		actualURI := MustURL(t, "example://example.com/123")

		c.Publish(actualURI, now.Add(1))
		select {
		case <-ctx.Done():
			t.Fatalf("expected to receive message, got timeout")
		case msg := <-ch:
			if msg == nil {
				t.Fatalf("expected non-nil message, got nil")
			}
			if msg.String() != actualURI.String() {
				t.Fatalf("expected %v, got %v", actualURI, msg)
			}
		}
	})
	t.Run("uri unmatched", func(t *testing.T) {
		ch := make(chan *url.URL, 1)
		defer close(ch)
		subscribeURI := MustURL(t, "example://example.com/{id}")
		now := MustTime(t, "2023-10-01T00:00:00Z")

		c := resourceChangeContext{
			ctx: t.Context(),
			subscriber: map[string]ResourceChangeSubscriber{
				"1": &resourceChangeSubscriber{
					id:            "1",
					subscribedURI: subscribeURI,
					lastReceived:  now,
					ch:            ch,
					nowFunc: func() time.Time {
						return now.Add(1)
					},
				},
			},
		}

		c.Publish(MustURL(t, "example://example.com"), now.Add(1))
		select {
		case <-ch:
			t.Fatalf("expected no message, got one")
		default:
			if len(ch) != 0 {
				t.Fatalf("expected no messages in channel, got %d", len(ch))
			}
		}
	})
	t.Run("received at is past", func(t *testing.T) {
		ch := make(chan *url.URL, 1)
		defer close(ch)
		subscribeURI := MustURL(t, "example://example.com/{id}")
		now := MustTime(t, "2023-10-01T00:00:00Z")

		c := resourceChangeContext{
			ctx: t.Context(),
			subscriber: map[string]ResourceChangeSubscriber{
				"1": &resourceChangeSubscriber{
					id:            "1",
					subscribedURI: subscribeURI,
					lastReceived:  now,
					ch:            ch,
					nowFunc: func() time.Time {
						return now.Add(1)
					},
				},
			},
		}

		c.Publish(MustURL(t, "example://example.com/123"), now.Add(-1))
		select {
		case <-ch:
			t.Fatalf("expected no message, got one")
		default:
			if len(ch) != 0 {
				t.Fatalf("expected no messages in channel, got %d", len(ch))
			}
		}
	})
}

func TestResourceChangeContext_subscribe(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		ch := make(chan *url.URL, 1)
		defer close(ch)
		uri := MustURL(t, "example://example.com")
		now := MustTime(t, "2023-10-01T00:00:00Z")

		c := resourceChangeContext{
			ctx: t.Context(),
		}

		subscriber := &resourceChangeSubscriber{
			id:            "1",
			subscribedURI: uri,
			lastReceived:  now,
			ch:            ch,
			nowFunc: func() time.Time {
				return now.Add(1)
			},
		}

		c.subscribe(subscriber)
		if len(c.subscriber) != 1 {
			t.Fatalf("expected 1 subscriber, got %d", len(c.subscriber))
		}
		if !reflect.DeepEqual(c.subscriber["1"], subscriber) {
			t.Fatalf("expected %v, got %v", subscriber, c.subscriber["1"])
		}
	})
}

func TestResourceChangeContext_unsubscribe(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		ch := make(chan *url.URL, 1)
		defer close(ch)
		uri := MustURL(t, "example://example.com")
		now := MustTime(t, "2023-10-01T00:00:00Z")

		c := resourceChangeContext{
			ctx: t.Context(),
			subscriber: map[string]ResourceChangeSubscriber{
				"1": &resourceChangeSubscriber{
					id:            "1",
					subscribedURI: uri,
					lastReceived:  now,
					ch:            ch,
					nowFunc: func() time.Time {
						return now.Add(1)
					},
				},
			},
		}
		c.unsubscribe("1")
		if len(c.subscriber) != 0 {
			t.Fatalf("expected 0 subscribers, got %d", len(c.subscriber))
		}
	})
}

func Test_uriMatches(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		subscribeURI := MustURL(t, "example://example.com/{id}")
		actualURI := MustURL(t, "example://example.com/123")
		if !uriMatches(actualURI, subscribeURI) {
			t.Fatalf("expected true, got false")
		}
	})
	t.Run("no match", func(t *testing.T) {
		subscribeURI := MustURL(t, "example://example.com/{id}")
		actualURI := MustURL(t, "example://example.com")
		if uriMatches(actualURI, subscribeURI) {
			t.Fatalf("expected false, got true")
		}
	})
	t.Run("no path params", func(t *testing.T) {
		subscribeURI := MustURL(t, "example://example.com")
		actualURI := MustURL(t, "example://example.com")
		if !uriMatches(actualURI, subscribeURI) {
			t.Fatalf("expected true, got false")
		}
	})
	t.Run("different scheme", func(t *testing.T) {
		subscribeURI := MustURL(t, "http://example.com")
		actualURI := MustURL(t, "example://example.com")
		if uriMatches(actualURI, subscribeURI) {
			t.Fatalf("expected false, got true")
		}
	})
	t.Run("different host", func(t *testing.T) {
		subscribeURI := MustURL(t, "example://example.com")
		actualURI := MustURL(t, "example://example.org")
		if uriMatches(actualURI, subscribeURI) {
			t.Fatalf("expected false, got true")
		}
	})
	t.Run("different path", func(t *testing.T) {
		subscribeURI := MustURL(t, "example://example.com/test")
		actualURI := MustURL(t, "example://example.com/123")
		if uriMatches(actualURI, subscribeURI) {
			t.Fatalf("expected false, got true")
		}
	})
	t.Run("different depth of path", func(t *testing.T) {
		subscribeURI := MustURL(t, "example://example.com/test/123")
		actualURI := MustURL(t, "example://example.com/123")
		if uriMatches(actualURI, subscribeURI) {
			t.Fatalf("expected false, got true")
		}
	})
	t.Run("nil uri", func(t *testing.T) {
		subscribeURI := MustURL(t, "example://example.com")
		if uriMatches(subscribeURI, nil) {
			t.Fatalf("expected false, got true")
		}
	})
}

func TestResourceListChangeSubscriber_ID(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		s := resourceListChangeSubscriber{
			id: "test-id",
		}
		if got := s.ID(); got != "test-id" {
			t.Fatalf("expected 'test-id', got %v", got)
		}
	})
}

func TestResourceListChangeSubscriber_LastReceived(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		time := MustTime(t, "2023-10-01T00:00:00Z")
		s := resourceListChangeSubscriber{
			lastReceived: time,
		}
		if got := s.LastReceived(); !got.Equal(time) {
			t.Fatalf("expected %v, got %v", time, got)
		}
	})
	t.Run("unset", func(t *testing.T) {
		time := MustTime(t, "0001-01-01T00:00:00Z")
		s := resourceListChangeSubscriber{}
		if got := s.LastReceived(); !got.Equal(time) {
			t.Fatalf("expected %v, got %v", time, got)
		}
	})
}

func TestResourceListChangeSubscriber_Publish(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		ch := make(chan struct{}, 1)
		defer close(ch)
		now := MustTime(t, "2023-10-01T00:00:00Z")
		s := resourceListChangeSubscriber{
			id:           "test-id",
			lastReceived: now,
			ch:           ch,
			nowFunc: func() time.Time {
				return now.Add(1)
			},
		}

		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()

		s.Publish()
		select {
		case <-ctx.Done():
			t.Fatalf("expected to receive message, got timeout")
		case <-ch:
			if !s.lastReceived.Equal(now.Add(1)) {
				t.Fatalf("expected %v, got %v", now.Add(1), s.lastReceived)
			}
		}
	})
}

func TestResourceListChangeSubscriber_reset(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		ch := make(chan struct{}, 1)
		now := MustTime(t, "2023-10-01T00:00:00Z")
		s := resourceListChangeSubscriber{
			id:           "test-id",
			lastReceived: now,
			ch:           ch,
			nowFunc: func() time.Time {
				return now.Add(1)
			},
		}
		s.reset()
		if s.id != "" {
			t.Fatalf("expected empty string, got %v", s.id)
		}
		if s.lastReceived != (time.Time{}) {
			t.Fatalf("expected zero time, got %v", s.lastReceived)
		}
		if s.ch != nil {
			t.Fatalf("expected nil, got %v", s.ch)
		}
		if s.nowFunc == nil {
			t.Fatalf("expected non-nil, got nil")
		}
	})
}

func TestResourceListChangeContext_Context(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		c := resourceListChangeContext{
			ctx: t.Context(),
		}
		if got := c.Context(); got != t.Context() {
			t.Fatalf("expected %v, got %v", t.Context(), got)
		}
	})
}

func TestResourceListChangeContext_Publish(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		ch := make(chan struct{}, 1)
		defer close(ch)
		now := MustTime(t, "2023-10-01T00:00:00Z")
		subscriber := &resourceListChangeSubscriber{
			id:           "test-id",
			lastReceived: now,
			ch:           ch,
			nowFunc: func() time.Time {
				return now.Add(1)
			},
		}

		c := resourceListChangeContext{
			ctx: t.Context(),
			subscriber: map[string]ResourceListChangeSubscriber{
				subscriber.ID(): subscriber,
			},
		}

		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()

		c.Publish(now.Add(1))
		select {
		case <-ctx.Done():
			t.Fatalf("expected to receive message, got timeout")
		case <-ch:
			if !subscriber.lastReceived.Equal(now.Add(1)) {
				t.Fatalf("expected %v, got %v", now.Add(1), subscriber.lastReceived)
			}
		}
	})
	t.Run("received at is past", func(t *testing.T) {
		ch := make(chan struct{}, 1)
		defer close(ch)
		now := MustTime(t, "2023-10-01T00:00:00Z")
		subscriber := &resourceListChangeSubscriber{
			id:           "test-id",
			lastReceived: now,
			ch:           ch,
			nowFunc: func() time.Time {
				return now.Add(1)
			},
		}

		c := resourceListChangeContext{
			ctx: t.Context(),
			subscriber: map[string]ResourceListChangeSubscriber{
				subscriber.ID(): subscriber,
			},
		}

		c.Publish(now.Add(-1))
		select {
		case <-ch:
			t.Fatalf("expected no message, got message")
		default:
		}
	})
}

func TestResourceListChangeContext_subscribe(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		ch := make(chan struct{}, 1)
		defer close(ch)
		now := MustTime(t, "2023-10-01T00:00:00Z")
		subscriber := &resourceListChangeSubscriber{
			id:           "test-id",
			lastReceived: now,
			ch:           ch,
			nowFunc: func() time.Time {
				return now.Add(1)
			},
		}

		c := resourceListChangeContext{
			ctx: t.Context(),
		}

		c.subscribe(subscriber)
		if len(c.subscriber) != 1 {
			t.Fatalf("expected 1 subscriber, got %d", len(c.subscriber))
		}
		if !reflect.DeepEqual(c.subscriber[subscriber.ID()], subscriber) {
			t.Fatalf("expected %v, got %v", subscriber, c.subscriber[subscriber.ID()])
		}
	})
}

func TestResourceListChangeContext_unsubscribe(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		ch := make(chan struct{}, 1)
		defer close(ch)
		now := MustTime(t, "2023-10-01T00:00:00Z")
		subscriber := &resourceListChangeSubscriber{
			id:           "test-id",
			lastReceived: now,
			ch:           ch,
			nowFunc: func() time.Time {
				return now.Add(1)
			},
		}

		c := resourceListChangeContext{
			ctx: t.Context(),
			subscriber: map[string]ResourceListChangeSubscriber{
				subscriber.ID(): subscriber,
			},
		}
		c.unsubscribe(subscriber.ID())
		if len(c.subscriber) != 0 {
			t.Fatalf("expected 0 subscribers, got %d", len(c.subscriber))
		}
	})
}

func TestPromptContext_PromptName(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		c := newPromptContext(nil, nil, nil)
		c.promptName = "test-prompt"
		if got := c.PromptName(); got != "test-prompt" {
			t.Fatalf("expected 'test-prompt', got %v", got)
		}
	})
	t.Run("unset", func(t *testing.T) {
		c := newPromptContext(nil, nil, nil)
		if got := c.PromptName(); got != "" {
			t.Fatalf("expected empty string, got %v", got)
		}
	})
}

func TestPromptContext_Arguments(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		args := map[string]any{"name": "John", "age": 30}
		c := newPromptContext(nil, nil, nil)
		c.args = args
		if got := c.Arguments(); !reflect.DeepEqual(got, args) {
			t.Fatalf("expected %v, got %v", args, got)
		}
	})
	t.Run("unset", func(t *testing.T) {
		c := newPromptContext(nil, nil, nil)
		if got := c.Arguments(); got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})
}

func TestPromptContext_Bind(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		type Args struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		c := newPromptContext(json.Unmarshal, json.Marshal, nil)
		c.args = map[string]any{"name": "John", "age": 30}
		var args Args
		err := c.Bind(&args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if args.Name != "John" || args.Age != 30 {
			t.Fatalf("expected name=John, age=30, got name=%v, age=%v", args.Name, args.Age)
		}
	})
	t.Run("empty args", func(t *testing.T) {
		type Args struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		c := newPromptContext(json.Unmarshal, json.Marshal, nil)
		var args Args
		err := c.Bind(&args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if args.Name != "" || args.Age != 0 {
			t.Fatalf("expected name='', age=0, got name=%v, age=%v", args.Name, args.Age)
		}
	})
}

func TestPromptContext_String(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		var dest getPromptResult
		c := newPromptContext(nil, json.Marshal, nil)
		c.dest = &dest
		if err := c.String("user", "Hello world"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(c.dest.Messages) != 1 {
			t.Fatalf("expected 1 message, got %d", len(c.dest.Messages))
		}
		msg := c.dest.Messages[0]
		if msg.Role != "user" {
			t.Fatalf("expected role 'user', got %v", msg.Role)
		}
		textContent, ok := msg.Content.(*textPromptContent)
		if !ok {
			t.Fatalf("expected *textPromptContent, got %T", msg.Content)
		}
		if textContent.Text != "Hello world" {
			t.Fatalf("expected 'Hello world', got %v", textContent.Text)
		}
	})
}

func TestPromptContext_JSON(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		type Data struct {
			Message string `json:"message"`
		}
		var dest getPromptResult
		c := newPromptContext(nil, json.Marshal, nil)
		c.dest = &dest
		data := Data{Message: "hello"}
		if err := c.JSON("assistant", data); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(c.dest.Messages) != 1 {
			t.Fatalf("expected 1 message, got %d", len(c.dest.Messages))
		}
		msg := c.dest.Messages[0]
		if msg.Role != "assistant" {
			t.Fatalf("expected role 'assistant', got %v", msg.Role)
		}
		textContent, ok := msg.Content.(*textPromptContent)
		if !ok {
			t.Fatalf("expected *textPromptContent, got %T", msg.Content)
		}
		expected := `{"message":"hello"}`
		if textContent.Text != expected {
			t.Fatalf("expected %s, got %v", expected, textContent.Text)
		}
	})
}

func TestPromptContext_Image(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		var dest getPromptResult
		c := newPromptContext(nil, json.Marshal, func(data []byte) string {
			return "encoded-image-data"
		})
		c.dest = &dest
		imageData := []byte("image-data")
		if err := c.Image("user", imageData, "image/png"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(c.dest.Messages) != 1 {
			t.Fatalf("expected 1 message, got %d", len(c.dest.Messages))
		}
		msg := c.dest.Messages[0]
		if msg.Role != "user" {
			t.Fatalf("expected role 'user', got %v", msg.Role)
		}
		imageContent, ok := msg.Content.(*imagePromptContent)
		if !ok {
			t.Fatalf("expected *imagePromptContent, got %T", msg.Content)
		}
		if imageContent.Data != "encoded-image-data" {
			t.Fatalf("expected 'encoded-image-data', got %v", imageContent.Data)
		}
		if imageContent.MimeType != "image/png" {
			t.Fatalf("expected 'image/png', got %v", imageContent.MimeType)
		}
	})
}

func TestPromptContext_Resource(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		var dest getPromptResult
		c := newPromptContext(nil, json.Marshal, nil)
		c.dest = &dest
		uri := MustURL(t, "example://example.com")
		resource := &textResourceContent{
			resourceContentBase: resourceContentBase{
				uri:      weak.Make(uri),
				mimeType: "text/plain",
			},
			text:    "resource content",
			marshal: json.Marshal,
		}
		if err := c.Resource("system", resource); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(c.dest.Messages) != 1 {
			t.Fatalf("expected 1 message, got %d", len(c.dest.Messages))
		}
		msg := c.dest.Messages[0]
		if msg.Role != "system" {
			t.Fatalf("expected role 'system', got %v", msg.Role)
		}
		resourceContent, ok := msg.Content.(*embedResourcePromptContent)
		if !ok {
			t.Fatalf("expected *embedResourcePromptContent, got %T", msg.Content)
		}
		if !reflect.DeepEqual(resourceContent.Resource, resource) {
			t.Fatalf("expected %v, got %v", resource, resourceContent.Resource)
		}
	})
}

func TestPromptContext_reset(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		var dest getPromptResult
		c := newPromptContext(json.Unmarshal, json.Marshal, base64.StdEncoding.EncodeToString)
		c.promptName = "test-prompt"
		c.args = map[string]any{"name": "John"}
		c.dest = &dest
		c.jsonrpcRequest = &jsonrpc2.Request{}
		c.store.Store("key", "value")
		c.reset()
		if c.promptName != "" {
			t.Fatalf("expected empty string, got %v", c.promptName)
		}
		if c.args != nil {
			t.Fatalf("expected nil, got %v", c.args)
		}
		if c.dest != nil {
			t.Fatalf("expected nil, got %v", c.dest)
		}
		if c.jsonrpcRequest != nil {
			t.Fatalf("expected nil, got %v", c.jsonrpcRequest)
		}
		c.store.Range(func(k, v interface{}) bool {
			t.Fatalf("unexpected key=%v value=%v", k, v)
			return true
		})
	})
}
