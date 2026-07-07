// MCP server mode (`wails3 translate --mcp`): a stdio Model Context Protocol
// server so any MCP client harness can translate the docs using its own model,
// while this server provides the deterministic parts - staleness, per-file
// translation context, and a validating write that stamps the cache and queues
// files for human review. See MCP_SPEC.md.
//
// The JSON-RPC subset is copy-adapted from the app-control MCP server
// (v3/pkg/application/mcp_protocol_enabled.go); this one is stdio instead of
// HTTP and ships unconditionally (no build tag) because it is a CLI feature.
package i18n

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// mcpProtocolVersion is the latest MCP protocol revision this server supports.
// Older revisions are accepted; the subset we implement (initialize, ping,
// tools/list, tools/call) is identical across them.
const mcpProtocolVersion = "2025-06-18"

const mcpServerVersion = "1.0.0"

// JSON-RPC 2.0 error codes.
const (
	mcpCodeParseError     = -32700
	mcpCodeInvalidRequest = -32600
	mcpCodeMethodNotFound = -32601
	mcpCodeInvalidParams  = -32602
)

// mcpInstructions is pulled into the client's context on initialize - it is how
// a connecting agent learns the workflow without any external skill or doc.
const mcpInstructions = "Translation server for the Wails documentation (Astro Starlight). " +
	"Workflow: call translation_status to pick a locale, list_files (filter 'needs_work') " +
	"for the work queue, then for each file: get_translation_context, translate the source " +
	"into the target language yourself, and submit the complete translated file with " +
	"write_translation. Rules: translate prose only - never translate code blocks, inline " +
	"code, import lines, JSX/MDX tags, or glossary terms; keep frontmatter keys and " +
	"translate only the title and description values; preserve document structure (same " +
	"headings, fences, components). For stale files, use the drift blocks to update only " +
	"what changed in the source. Writes are validated, cache-stamped, and queued as " +
	"needs_review for human review in the wails3 translate workbench. If no docs site is " +
	"found, call setup_docs to clone the Wails docs."

type mcpJSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type mcpJSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type mcpJSONRPCResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      json.RawMessage  `json:"id"`
	Result  any              `json:"result,omitempty"`
	Error   *mcpJSONRPCError `json:"error,omitempty"`
}

func mcpErrorResponse(id json.RawMessage, code int, format string, args ...any) *mcpJSONRPCResponse {
	return &mcpJSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &mcpJSONRPCError{Code: code, Message: fmt.Sprintf(format, args...)},
	}
}

func mcpResultResponse(id json.RawMessage, result any) *mcpJSONRPCResponse {
	return &mcpJSONRPCResponse{JSONRPC: "2.0", ID: id, Result: result}
}

// mcpTool is one callable tool: schema for tools/list, handler for tools/call.
type mcpTool struct {
	Name        string
	Description string
	Schema      map[string]any
	Handler     func(args map[string]any) (any, error)
}

// MCPOptions configure an MCP server session.
type MCPOptions struct {
	Dir string // docs root override; auto-detected when empty
}

// mcpServer holds the (lazily initialized) translation session and tool set.
type mcpServer struct {
	dir     string         // --dir override, kept for setup_docs
	session *helperSession // nil until a docs root is resolved
	tools   []*mcpTool
	logw    io.Writer // diagnostics; never stdout (that carries the protocol)
}

// RunMCP serves MCP over stdio until the client closes stdin. It never prompts
// and writes nothing to stdout except JSON-RPC messages.
func RunMCP(opts MCPOptions) error {
	return newMCPServer(opts).serve(os.Stdin, os.Stdout)
}

func newMCPServer(opts MCPOptions) *mcpServer {
	s := &mcpServer{dir: opts.Dir, logw: os.Stderr}
	// Best-effort eager session; when nothing is found the server still starts
	// and every tool except setup_docs answers "call setup_docs first".
	if sess, err := newHelperSession(opts.Dir); err == nil {
		s.session = sess
	}
	s.registerTools()
	return s
}

// sess gates tools that need a resolved docs root.
func (s *mcpServer) sess() (*helperSession, error) {
	if s.session == nil {
		return nil, errors.New("no docs site found; call setup_docs first")
	}
	return s.session, nil
}

// serve is the sequential request loop. json.Decoder (not bufio.Scanner, whose
// default 64KB token limit a single write_translation request can exceed) reads
// one message per iteration; json.Encoder emits exactly one \n-terminated line
// per response, which is the MCP stdio framing.
func (s *mcpServer) serve(r io.Reader, w io.Writer) error {
	dec := json.NewDecoder(r)
	enc := json.NewEncoder(w)
	for {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			if errors.Is(err, io.EOF) {
				return nil // client closed stdin: clean shutdown
			}
			// A syntax error leaves the stream unrecoverable mid-message;
			// report it and stop rather than resync-guessing.
			_ = enc.Encode(mcpErrorResponse(nil, mcpCodeParseError, "parse error: %v", err))
			return fmt.Errorf("mcp: malformed input: %w", err)
		}

		requests, batch, err := mcpParseMessages(raw)
		if err != nil {
			if encErr := enc.Encode(mcpErrorResponse(nil, mcpCodeParseError, "parse error: %v", err)); encErr != nil {
				return encErr
			}
			continue
		}
		if batch && len(requests) == 0 {
			if encErr := enc.Encode(mcpErrorResponse(nil, mcpCodeInvalidRequest, "invalid request: empty batch")); encErr != nil {
				return encErr
			}
			continue
		}

		var responses []*mcpJSONRPCResponse
		for _, req := range requests {
			if response := s.handleMessage(req); response != nil {
				responses = append(responses, response)
			}
		}
		if len(responses) == 0 {
			continue // notifications only
		}
		var out any = responses[0]
		if batch {
			out = responses
		}
		if err := enc.Encode(out); err != nil {
			return err
		}
	}
}

// mcpParseMessages splits a raw message into requests, tolerating JSON-RPC
// batches (2025-03-26 clients may send them; 2025-06-18 removed batching).
func mcpParseMessages(body []byte) (requests []*mcpJSONRPCRequest, batch bool, err error) {
	for _, b := range body {
		if b == ' ' || b == '\t' || b == '\r' || b == '\n' {
			continue
		}
		batch = b == '['
		break
	}
	if batch {
		var reqs []*mcpJSONRPCRequest
		if err := json.Unmarshal(body, &reqs); err != nil {
			return nil, true, err
		}
		return reqs, true, nil
	}
	var req mcpJSONRPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, false, err
	}
	return []*mcpJSONRPCRequest{&req}, false, nil
}

// handleMessage processes a single JSON-RPC message. Returns nil for
// notifications, which receive no response.
func (s *mcpServer) handleMessage(req *mcpJSONRPCRequest) *mcpJSONRPCResponse {
	isNotification := len(req.ID) == 0 || string(req.ID) == "null"
	if isNotification || strings.HasPrefix(req.Method, "notifications/") {
		return nil
	}

	switch req.Method {
	case "initialize":
		var params struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		_ = json.Unmarshal(req.Params, &params)
		version := mcpProtocolVersion
		switch params.ProtocolVersion {
		case "2024-11-05", "2025-03-26":
			version = params.ProtocolVersion
		}
		return mcpResultResponse(req.ID, map[string]any{
			"protocolVersion": version,
			"capabilities": map[string]any{
				"tools": map[string]any{"listChanged": false},
			},
			"serverInfo": map[string]any{
				"name":    "wails-translate",
				"title":   "Wails Docs Translation",
				"version": mcpServerVersion,
			},
			"instructions": mcpInstructions,
		})

	case "ping":
		return mcpResultResponse(req.ID, map[string]any{})

	case "tools/list":
		tools := make([]map[string]any, 0, len(s.tools))
		for _, t := range s.tools {
			tools = append(tools, map[string]any{
				"name":        t.Name,
				"description": t.Description,
				"inputSchema": t.Schema,
			})
		}
		return mcpResultResponse(req.ID, map[string]any{"tools": tools})

	case "tools/call":
		return s.handleToolCall(req)

	default:
		return mcpErrorResponse(req.ID, mcpCodeMethodNotFound, "method not found: %s", req.Method)
	}
}

func (s *mcpServer) handleToolCall(req *mcpJSONRPCRequest) (response *mcpJSONRPCResponse) {
	var params struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return mcpErrorResponse(req.ID, mcpCodeInvalidParams, "invalid params: %v", err)
	}

	var selected *mcpTool
	for _, t := range s.tools {
		if t.Name == params.Name {
			selected = t
			break
		}
	}
	if selected == nil {
		return mcpErrorResponse(req.ID, mcpCodeInvalidParams, "unknown tool: %s", params.Name)
	}
	if params.Arguments == nil {
		params.Arguments = map[string]any{}
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			fmt.Fprintf(s.logw, "mcp: tool %s panicked: %v\n", params.Name, recovered)
			response = mcpResultResponse(req.ID, mcpToolError(fmt.Sprintf("tool %s panicked: %v", params.Name, recovered)))
		}
	}()

	result, err := selected.Handler(params.Arguments)
	if err != nil {
		// Tool failures are agent-visible results (retryable), not protocol errors.
		return mcpResultResponse(req.ID, mcpToolError(err.Error()))
	}
	return mcpResultResponse(req.ID, mcpToolResult(result))
}

// mcpToolResult converts a tool's return value into an MCP CallToolResult.
func mcpToolResult(value any) map[string]any {
	var text string
	switch v := value.(type) {
	case nil:
		text = "ok"
	case string:
		text = v
	default:
		data, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			text = fmt.Sprintf("%v", v)
		} else {
			text = string(data)
		}
	}
	return map[string]any{
		"content": []map[string]any{{"type": "text", "text": text}},
		"isError": false,
	}
}

func mcpToolError(message string) map[string]any {
	return map[string]any{
		"content": []map[string]any{{"type": "text", "text": message}},
		"isError": true,
	}
}
