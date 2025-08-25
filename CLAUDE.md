# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based MCP (Model Context Protocol) server that provides home automation control through Aqara/Xiaomi smart home devices. The server exposes MCP tools for listing and controlling device buttons/scenes in a smart home setup.

## Commands

### Development Commands
- `go run .` - Run the server locally
- `go build .` - Build the binary
- `go mod tidy` - Clean up module dependencies
- `go mod download` - Download dependencies

### Server Configuration
The server runs on HTTP with configurable host/port via environment variables:
- Default: `127.0.0.1:8080`
- Override with `host` and `port` environment variables

## Architecture

### Core Components

**main.go** - HTTP server entry point
- Sets up MCP server with SSE (Server-Sent Events) transport
- Handles CORS and bearer token authentication
- Registers MCP tools and middleware
- Configurable host/port from environment

**service.go** - MCP tool implementations and API wrappers
- `list_device_control_buttons` - Lists available smart home control buttons
- `push_device_control_button` - Executes device control commands
- Wraps Aqara cloud API calls with proper error handling
- Contains device/room mapping notes in Chinese

**smh.go** - HTTP client and Aqara API integration
- Handles authenticated API calls to Aqara cloud service
- Implements request signing with HMAC-SHA256
- Provides generic service calling with JSON marshaling
- Device ID generation from MAC address/hostname

### Authentication & Security
- Bearer token authentication via `API_TOKEN` environment variable
- Request signing using `AppID`, `AppSecret` and HMAC-SHA256
- Device fingerprinting for API identification
- CORS enabled for web client access

### MCP Integration
- Uses `github.com/modelcontextprotocol/go-sdk/mcp` for MCP protocol
- Exposes tools as MCP functions with JSON schema validation
- SSE transport for real-time communication
- Request/response logging middleware

### Configuration Requirements
Environment variables needed in `.env`:
- `API_KEY` - Aqara cloud service API key
- `API_TOKEN` - Authentication token for MCP clients

### Smart Home Domain
The service specifically handles Chinese smart home scenarios with room types:
- 客厅 (Living room), 厨房 (Kitchen), 玄关 (Entrance)
- 主卧 (Master bedroom), 次卧 (Secondary bedroom), 卫生间 (Bathroom)
- 走廊 (Corridor) connecting all rooms
- Various lighting controls and scene automation

## Key Dependencies
- `github.com/modelcontextprotocol/go-sdk` - MCP protocol implementation
- `github.com/devfans/golang/log` - Structured logging
- `github.com/devfans/envconf/dotenv` - Environment configuration
- `github.com/google/uuid` - UUID generation for requests