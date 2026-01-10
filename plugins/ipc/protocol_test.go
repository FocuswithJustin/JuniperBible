package ipc

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestReadRequest(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Request
		wantErr bool
	}{
		{
			name:  "valid request",
			input: `{"command":"detect","args":{"path":"/test"}}`,
			want: &Request{
				Command: "detect",
				Args:    map[string]interface{}{"path": "/test"},
			},
		},
		{
			name:  "empty args",
			input: `{"command":"info"}`,
			want:  &Request{Command: "info"},
		},
		{
			name:    "invalid json",
			input:   `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replace stdin temporarily
			oldStdin := os.Stdin
			r, w, _ := os.Pipe()
			os.Stdin = r
			w.WriteString(tt.input)
			w.Close()
			defer func() { os.Stdin = oldStdin }()

			got, err := ReadRequest()
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Command != tt.want.Command {
				t.Errorf("Command = %v, want %v", got.Command, tt.want.Command)
			}
		})
	}
}

func TestRespond(t *testing.T) {
	tests := []struct {
		name   string
		result interface{}
	}{
		{
			name:   "detect result",
			result: &DetectResult{Detected: true, Format: "zip"},
		},
		{
			name: "ingest result",
			result: &IngestResult{
				ArtifactID: "test",
				BlobSHA256: "abc123",
				SizeBytes:  100,
			},
		},
		{
			name:   "nil result",
			result: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := Respond(tt.result)
			w.Close()
			os.Stdout = oldStdout

			if err != nil {
				t.Fatalf("Respond() error = %v", err)
			}

			var buf bytes.Buffer
			buf.ReadFrom(r)

			var resp Response
			if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			if resp.Status != "ok" {
				t.Errorf("Status = %v, want ok", resp.Status)
			}
		})
	}
}

func TestMustRespond(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	MustRespond(&DetectResult{Detected: true})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var resp Response
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("Status = %v, want ok", resp.Status)
	}
}

func TestReadRequestEmptyInput(t *testing.T) {
	// Replace stdin with empty pipe
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	w.Close() // Close immediately for EOF
	defer func() { os.Stdin = oldStdin }()

	_, err := ReadRequest()
	if err == nil {
		t.Error("ReadRequest() expected error for empty input, got nil")
	}
}

func TestReadRequestTruncatedJSON(t *testing.T) {
	// Truncated JSON input
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	w.WriteString(`{"command":"detect","args":`)
	w.Close()
	defer func() { os.Stdin = oldStdin }()

	_, err := ReadRequest()
	if err == nil {
		t.Error("ReadRequest() expected error for truncated JSON, got nil")
	}
}

func TestRespondEnumerateResult(t *testing.T) {
	result := &EnumerateResult{
		Entries: []EnumerateEntry{
			{
				Path:      "file1.txt",
				SizeBytes: 100,
				IsDir:     false,
				ModTime:   "2024-01-01T00:00:00Z",
				Metadata:  map[string]string{"type": "text"},
			},
			{
				Path:      "subdir",
				SizeBytes: 0,
				IsDir:     true,
			},
		},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := Respond(result)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Respond() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var resp Response
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("Status = %v, want ok", resp.Status)
	}

	// Verify entries are present
	resultMap, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatalf("Result is not a map")
	}
	entries, ok := resultMap["entries"].([]interface{})
	if !ok || len(entries) != 2 {
		t.Errorf("entries = %v, want 2 entries", entries)
	}
}

func TestRespondExtractIRResult(t *testing.T) {
	result := &ExtractIRResult{
		IR: map[string]interface{}{
			"books": []string{"Genesis", "Exodus"},
		},
		LossClass: "L2",
		LossReport: &LossReport{
			SourceFormat: "Test",
			TargetFormat: "IR",
			LossClass:    "L2",
			LostElements: []LostElement{
				{ElementType: "footnote", Reason: "not supported"},
			},
			Warnings: []string{"some content simplified"},
		},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := Respond(result)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Respond() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var resp Response
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("Status = %v, want ok", resp.Status)
	}
}

func TestRespondEmitNativeResult(t *testing.T) {
	result := &EmitNativeResult{
		OutputPath: "output/file.xml",
		Format:     "XML",
		LossClass:  "L1",
		LossReport: &LossReport{
			SourceFormat: "IR",
			TargetFormat: "XML",
			LossClass:    "L1",
			Warnings:     []string{"simplified formatting"},
		},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := Respond(result)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Respond() error = %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var resp Response
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("Status = %v, want ok", resp.Status)
	}
}

func TestRespondUnencodableResult(t *testing.T) {
	// channels are not JSON-encodable
	ch := make(chan int)
	result := map[string]interface{}{
		"channel": ch,
	}

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := Respond(result)
	w.Close()
	os.Stdout = oldStdout

	if err == nil {
		t.Error("Respond() expected error for unencodable result, got nil")
	}
}

func TestRequestStructure(t *testing.T) {
	// Test that Request can handle various command types
	inputs := []string{
		`{"command":"detect","args":{"path":"/test"}}`,
		`{"command":"ingest","args":{"path":"/test","output_dir":"/out"}}`,
		`{"command":"enumerate","args":{"path":"/test"}}`,
		`{"command":"extract_ir","args":{"path":"/test","output_dir":"/out"}}`,
		`{"command":"emit","args":{"ir":{},"output_dir":"/out"}}`,
		`{"command":"info"}`,
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			var req Request
			if err := json.Unmarshal([]byte(input), &req); err != nil {
				t.Errorf("failed to parse %q: %v", input, err)
			}
		})
	}
}

func TestResponseStructure(t *testing.T) {
	// Test that Response serializes correctly
	tests := []struct {
		name     string
		response Response
	}{
		{
			name:     "success with result",
			response: Response{Status: "ok", Result: map[string]string{"key": "value"}},
		},
		{
			name:     "error",
			response: Response{Status: "error", Error: "something went wrong"},
		},
		{
			name:     "success nil result",
			response: Response{Status: "ok"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.response)
			if err != nil {
				t.Errorf("failed to marshal: %v", err)
			}

			var decoded Response
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Errorf("failed to unmarshal: %v", err)
			}

			if decoded.Status != tt.response.Status {
				t.Errorf("Status = %v, want %v", decoded.Status, tt.response.Status)
			}
		})
	}
}
