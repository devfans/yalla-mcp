# Yalla MCP Server

A Go-based Model Context Protocol (MCP) server for controlling Aqara smart home devices. This server provides MCP tools for listing and controlling device buttons/scenes in a smart home setup.

## Features

- **MCP Integration**: Exposes smart home controls as MCP tools
- **Aqara Cloud API**: Integrates with Aqara cloud service for device management
- **HTTP Transport**: Runs as HTTP server with Server-Sent Events (SSE)
- **Authentication**: Bearer token authentication with request signing
- **CORS Support**: Web client compatible

## Quick Start

### Prerequisites

- Go 1.24.5 or later
- Aqara cloud service account and API credentials

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd yalla-mcp
```

2. Install dependencies:
```bash
go mod download
```

3. Configure environment variables in `.env`:
```env
API_KEY=your_aqara_api_key
API_TOKEN=your_authentication_token
host=127.0.0.1
port=8080
```

### Running

```bash
# Development
go run .

# Build and run
go build .
./main
```

The server will start on `http://127.0.0.1:8080` by default.

## MCP Tools

### `list_device_control_buttons`

Lists all available device control buttons in the current home.

**Returns**: Control buttons information in Markdown format

### `push_device_control_button`

Executes a device control command by pushing a specific button.

**Parameters**:
- `button` (integer): The control button ID to push

**Returns**: Device control result message

## Smart Home Layout

The system is designed for Chinese smart home scenarios with the following room types:

- **客厅** (Living room) - includes desktop lighting and TV lighting strips
- **厨房** (Kitchen) 
- **玄关** (Entrance)
- **主卧** (Master bedroom) - with ceiling light, left light, and right light
- **次卧** (Secondary bedroom)
- **卫生间** (Bathroom)
- **走廊** (Corridor) - connects all rooms

### Room Control Notes

- Room-wide controls (e.g., "客厅打开") turn on all lights in that room
- Individual device controls are also available
- Corridor lighting is needed for dining (餐桌) scenarios along with kitchen lighting

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `API_KEY` | Aqara cloud service API key | Required |
| `API_TOKEN` | Authentication token for MCP clients | Required |
| `host` | Server bind address | `127.0.0.1` |
| `port` | Server port | `8080` |

### Authentication

The server uses two-layer security:

1. **Bearer Token**: HTTP Authorization header with `API_TOKEN`
2. **Request Signing**: HMAC-SHA256 signing of requests to Aqara API

## API Integration

The server communicates with Aqara's cloud service at `https://ai-echo.aqara.cn/echo/mcp`. All requests are signed using:

- App ID generated from device fingerprint
- Secret key retrieved from Aqara service
- HMAC-SHA256 request signatures

## Development

### Project Structure

```
├── main.go     # HTTP server and MCP setup
├── service.go  # MCP tool implementations
├── smh.go      # Aqara API client and HTTP utilities
├── go.mod      # Go module dependencies
└── .env        # Environment configuration
```

### Logging

The server uses structured logging with appropriate log levels:
- `Debug`: HTTP requests and token verification
- `Info`: Tool calls and successful operations  
- `Warn`: Non-critical issues
- `Error`: Failed operations and API errors
