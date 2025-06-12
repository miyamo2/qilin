# Prompts Example

This example demonstrates how to use Qilin's MCP prompts functionality to create AI prompt templates with customizable parameters.

## Overview

The prompts example showcases four different types of prompt templates:

1. **Greeting Prompt** - Simple greeting with optional name parameter
2. **Email Prompt** - Professional email template with recipient, subject, and tone parameters
3. **Code Review Prompt** - Code review template with language and focus parameters
4. **Documentation Prompt** - Documentation template with type and audience parameters

## Features Demonstrated

- **Prompt Registration**: Using `q.Prompt()` to register prompt templates
- **Optional Arguments**: All prompts use optional arguments with sensible defaults
- **Multiple Parameters**: Email and code review prompts show multiple parameter usage
- **System/User Messages**: Prompts generate both system and user messages
- **Argument Validation**: Handler functions validate and provide defaults for arguments

## Running the Examples

### Stdio Transport
```bash
cd examples/prompts
go run cmd/prompts-stdio/main.go
```

### HTTP Streamable Transport
```bash
cd examples/prompts
go run cmd/prompts-streamable/main.go
```

## MCP Client Usage

Once running, you can use any MCP client to interact with the prompts:

### List Available Prompts
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "prompts/list"
}
```

### Get a Greeting Prompt
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "prompts/get",
  "params": {
    "name": "greeting",
    "arguments": {
      "name": "Alice"
    }
  }
}
```

### Get an Email Prompt with Parameters
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "prompts/get",
  "params": {
    "name": "email",
    "arguments": {
      "recipient": "John",
      "subject": "Project Update",
      "tone": "professional"
    }
  }
}
```

## Architecture

The example follows the same clean architecture pattern as other Qilin examples:

- `cmd/` - Entry points for stdio and streamable transports
- `handler/` - Prompt handler implementations
- `go.mod` - Module definition with local Qilin dependency

Each prompt handler receives a `qilin.PromptContext` which provides:
- `Arguments()` - Access to prompt arguments as a map
- `System(message)` - Add system messages to the prompt
- `User(message)` - Add user messages to the prompt