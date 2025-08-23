package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/devfans/envconf/dotenv"
	"github.com/devfans/golang/log"
	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)


var (
	host = dotenv.String("host", "127.0.0.1")
	port = dotenv.String("port", "8080")
)

func enableCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		if r.Method == "OPTIONS" {
			http.Error(w, "No Content", http.StatusNoContent)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

func verifyAuth(ctx context.Context, token string) (*auth.TokenInfo, error) {
	log.Debug("Token info", API_TOKEN, token)
	if token == API_TOKEN {
		return &auth.TokenInfo{
			Expiration: time.Now().Add(time.Hour * 24 * 365 * 10),
		}, nil
	}
	return nil, errors.New("invalid api key")
}

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
	if err := http.ListenAndServe(addr, enableCORS(auth.RequireBearerToken(verifyAuth, nil)(handler))); err != nil {
		log.Fatal("Failed to listen", "err", err)
	}
}
