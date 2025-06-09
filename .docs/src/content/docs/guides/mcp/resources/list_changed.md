---
title: Resource List Changed Notification
description: resource listChanged in qilin
---

Resource List Changed Notification is a feature of the Model Context Protocol (MCP) that allows clients to be notified when the list of available resources on your server changes. This is particularly useful for dynamic applications where resources can be added or removed at runtime.

## Registering a Resource List Change Observer

To register a resource list change observer in Qilin, you use the `ResourceListChangeObserver` method on your Qilin instance. This method takes a single parameter:

1. A handler function that monitors the resource list for changes.

```go
q.ResourceListChangeObserver(func(c qilin.ResourceListChangeContext) {
    // Handle resource list change notification
    return
})
```

## Publishing Resource List Changes

When your application detects that the resource list has changed (e.g., when resources are added or removed), you can notify clients by calling the `Publish` method on the `ResourceListChangeContext`:

```go /c.Publish/
func ResourceListChangeObserver(c qilin.ResourceListChangeContext) {
    // Monitor for changes to the resource list
    for t := range time.Tick(2 * time.Minute) {
        select {
        case <-c.Context().Done():
            return
        default:
            // When a change is detected, publish a notification
            c.Publish(t)
        }
    }
}
```

In this example:
- Set up a ticker that checks for changes every 2 minutes
- When a change is detected, we call `c.Publish()` with a timestamp to notify clients
