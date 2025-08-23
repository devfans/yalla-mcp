package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/devfans/envconf/dotenv"
	_ "github.com/devfans/envconf/dotenv"
	"github.com/devfans/golang/log"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)


var (
	host = dotenv.String("host", "127.0.0.1")
	port = dotenv.String("port", "8080")
)

func simpleResult(args ...string) *mcp.CallToolResult {
	contents := make([]mcp.Content, len(args))
	for i, v := range args {
		contents[i] =  &mcp.TextContent{Text: v} 
	}
	return &mcp.CallToolResult{
			Content: contents,
		}
}

func main() {
	loggingMiddleware := func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(
			ctx context.Context,
			method string,
			req mcp.Request,
		) (mcp.Result, error) {
			log.Info("MCP method started",
				"method", method,
				"session_id", req.GetSession().ID(),
				"has_params", req.GetParams() != nil,
			)
			// Log more for tool calls.
			if ctr, ok := req.(*mcp.CallToolRequest); ok {
				log.Info("Calling tool",
					"name", ctr.Params.Name,
					"args", ctr.Params.Arguments)
			}

			start := time.Now()
			result, err := next(ctx, method, req)
			duration := time.Since(start)
			if err != nil {
				log.Error("MCP method failed",
					"method", method,
					"session_id", req.GetSession().ID(),
					"duration_ms", duration.Milliseconds(),
					"err", err,
				)
			} else {
				log.Info("MCP method completed",
					"method", method,
					"session_id", req.GetSession().ID(),
					"duration_ms", duration.Milliseconds(),
					"has_result", result != nil,
				)
			}
			return result, err
		}
	}
	// Create a server with a single tool that says "Hi".
	server := mcp.NewServer(&mcp.Implementation{Name: "yalla"}, nil)
	server.AddReceivingMiddleware(loggingMiddleware)

	registerTools(server)

	// server.Run runs the server on the given transport.
	//
	// In this case, the server communicates over stdin/stdout.
	handler := mcp.NewSSEHandler(func(request *http.Request) *mcp.Server {
		return server
	})
	addr := fmt.Sprintf("%s:%s", host, port)
	log.Info("Server will start", "url", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal("Failed to listen", "err", err)
	}
}
