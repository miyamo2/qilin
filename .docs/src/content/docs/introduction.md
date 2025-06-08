---
title: Introduction
description: Qilin Introduction.
---

<img alt="Qilin Logo" src="https://raw.githubusercontent.com/miyamo2/qilin/refs/heads/main/.assets/logo.png" />

## What is Qilin?

Qilin is a Model Context Protocol (MCP) server framework for Go. It features a familiar look and feel, inspired by Go's well-known web application frameworks, making it easy for Go developers to build MCP-compliant servers.

## What is the Model Context Protocol (MCP)?

The [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) is a standardized protocol for communication between AI models and external tools or resources. It enables AI models to:

- Access external data and resources
- Call specialized tools and functions
- Maintain stateful sessions
- Receive notifications about resource changes

## Key Features of Qilin

### Easy-to-Use API

Qilin provides a clean, intuitive API that makes it simple to create MCP servers:

- Register resources with `q.Resource()`
- Register tools with `q.Tool()`
- Start the server with `q.Start()`

### Multiple Transport Options

Qilin supports different transport mechanisms:

- **Stdio Transport**: Simple communication via standard input/output
- **Streamable HTTP Transport**: Web-based communication over HTTP

### Resource Management

Qilin makes it easy to expose and manage resources:

- Static and dynamic resource URIs
- Resource change notifications
- Resource list management and updates

### Tool Integration

Easily integrate external tools and functions:

- JSON schema generation for request validation
- Support for various response types (JSON, text, images, etc.)
- Tool annotations for better client integration

### Session Management

Built-in session management capabilities:

- Automatic session creation and tracking
- Customizable session stores
- Session context for maintaining state

## Use Cases

Qilin is ideal for:

- Building AI assistants with access to specialized tools
- Creating interfaces between AI models and external data sources
- Developing custom MCP servers for specific domains
- Integrating AI capabilities into existing Go applications

## Getting Started

To get started with Qilin, check out the [Quick Start Guide](/qilin/guides/quickstart/) or explore the various guides in the documentation.
