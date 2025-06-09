---
title: Resource Subscriptions
description: resource subscribe in qilin
---

Resource Subscriptions is a feature of the Model Context Protocol (MCP) that allows clients to be notified when specific resources on your server change. This is particularly useful for applications that need to keep their data in sync with the server without constant polling.

## Registering a Resource Change Observer

To register a resource change observer in Qilin, you use the `ResourceChangeObserver` method on your Qilin instance. This method takes two parameters:

1. The URI pattern of the resources to observe (can include placeholders like `{id}`)
2. A handler function that monitors resource changes.

```go 
q.ResourceChangeObserver("example://example.com/{id}", func(c qilin.ResourceChangeContext) {
    // Handle resource change notification
    return
})
```

## Publishing Resource Changes

When your application detects that a resource has changed, you can notify clients by calling the `Publish` method on the `ResourceChangeContext`:

```go /c.Publish/
func WeatherForecastChangeObserver(c qilin.ResourceChangeContext) {
    for t := range time.Tick(time.Minute) {
        select {
        case <-c.Context().Done():
            return
        default:
            // When a resource changes, publish a notification
            uri, _ := url.Parse("weather://forecast/tokyo")
            c.Publish(uri, t)
        }
    }
}
```

In this example:
- Set up a ticker that checks for changes every minute
- When a change is detected, call `c.Publish()` with the specific URI of the changed resource and a timestamp
