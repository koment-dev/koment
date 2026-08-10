package lsp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func TestAFailedNotificationIsReportedAndTheSessionSurvives(t *testing.T) {
	var input bytes.Buffer
	writeTestMessage(t, &input, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{}})
	writeTestMessage(t, &input, map[string]any{
		"jsonrpc": "2.0", "method": "textDocument/didChange",
		"params": map[string]any{
			"textDocument":   map[string]any{"uri": "file:///nowhere.go", "version": 2},
			"contentChanges": []any{map[string]string{"text": "a"}, map[string]string{"text": "b"}},
		},
	})
	writeTestMessage(t, &input, map[string]any{"jsonrpc": "2.0", "id": 2, "method": "shutdown"})
	writeTestMessage(t, &input, map[string]any{"jsonrpc": "2.0", "method": "exit"})

	var output, stderr bytes.Buffer
	if err := Run(context.Background(), &input, &output, &stderr); err != nil {
		t.Fatalf("a failing notification ended the session: %v", err)
	}

	messages := readTestMessages(t, &output)
	reported := false
	for _, message := range messages {
		if message.Method != "window/showMessage" {
			continue
		}
		var params struct {
			Type    int    `json:"type"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(message.Params, &params); err != nil {
			t.Fatal(err)
		}
		if params.Type != messageTypeError {
			t.Errorf("reported with type %d, want %d", params.Type, messageTypeError)
		}
		if !strings.Contains(params.Message, "textDocument/didChange") {
			t.Errorf("the report does not name the method that failed: %q", params.Message)
		}
		reported = true
	}
	if !reported {
		t.Error("the failure never reached the client, so the editor shows no diagnostics and no reason")
	}
	if !strings.Contains(stderr.String(), "koment lsp:") {
		t.Error("the failure was not logged to stderr")
	}
	if len(responseByID(t, messages, "2").ID) == 0 {
		t.Error("the session did not answer shutdown, so it did not survive the failure")
	}
}

func TestASucceedingNotificationReportsNothing(t *testing.T) {
	var input bytes.Buffer
	writeTestMessage(t, &input, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{}})
	writeTestMessage(t, &input, map[string]any{
		"jsonrpc": "2.0", "method": "textDocument/didClose",
		"params": map[string]any{"textDocument": map[string]string{"uri": "file:///nowhere.go"}},
	})
	writeTestMessage(t, &input, map[string]any{"jsonrpc": "2.0", "method": "exit"})

	var output bytes.Buffer
	if err := Run(context.Background(), &input, &output, io.Discard); err != nil {
		t.Fatal(err)
	}
	for _, message := range readTestMessages(t, &output) {
		if message.Method == "window/showMessage" {
			t.Errorf("a healthy notification produced an error popup: %s", message.Params)
		}
	}
}
