package runner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Injectable functions for testing.
var (
	osCreate    = os.Create
	jsonMarshal = json.Marshal
	fileWrite   = func(w io.Writer, data []byte) (int, error) { return w.Write(data) }
	writeString = func(w io.StringWriter, s string) (int, error) { return w.WriteString(s) }
)

// TranscriptEvent represents a single event in a transcript JSONL file.
type TranscriptEvent struct {
	Type          string                 `json:"t"`
	Seq           int                    `json:"seq"`
	EngineID      string                 `json:"engine_id,omitempty"`
	PluginID      string                 `json:"plugin_id,omitempty"`
	PluginVersion string                 `json:"plugin_version,omitempty"`
	Module        string                 `json:"module,omitempty"`
	Key           string                 `json:"key,omitempty"`
	Profile       string                 `json:"profile,omitempty"`
	SHA256        string                 `json:"sha256,omitempty"`
	BLAKE3        string                 `json:"blake3,omitempty"`
	Bytes         int64                  `json:"bytes,omitempty"`
	Message       string                 `json:"message,omitempty"`
	Attributes    map[string]interface{} `json:"attributes,omitempty"`
}

// Known event types
const (
	EventEngineInfo       = "ENGINE_INFO"
	EventModuleDiscovered = "MODULE_DISCOVERED"
	EventKeyEnum          = "KEY_ENUM"
	EventEntryRendered    = "ENTRY_RENDERED"
	EventWarn             = "WARN"
	EventError            = "ERROR"
)

// ParseTranscript parses a transcript JSONL file and returns all events.
func ParseTranscript(path string) ([]TranscriptEvent, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open transcript: %w", err)
	}
	defer file.Close()

	var events []TranscriptEvent
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if line == "" {
			continue
		}

		var event TranscriptEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return nil, fmt.Errorf("failed to parse line %d: %w", lineNum, err)
		}

		events = append(events, event)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading transcript: %w", err)
	}

	return events, nil
}

// WriteTranscript writes a list of events to a transcript JSONL file.
func WriteTranscript(path string, events []TranscriptEvent) error {
	file, err := osCreate(path)
	if err != nil {
		return fmt.Errorf("failed to create transcript: %w", err)
	}
	defer file.Close()

	for _, event := range events {
		data, err := jsonMarshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}
		if _, err := fileWrite(file, data); err != nil {
			return fmt.Errorf("failed to write event: %w", err)
		}
		if _, err := writeString(file, "\n"); err != nil {
			return fmt.Errorf("failed to write newline: %w", err)
		}
	}

	return nil
}

// Transcript represents a parsed transcript with helper methods.
type Transcript struct {
	Events []TranscriptEvent
	Path   string
}

// LoadTranscript loads a transcript from a file.
func LoadTranscript(path string) (*Transcript, error) {
	events, err := ParseTranscript(path)
	if err != nil {
		return nil, err
	}
	return &Transcript{
		Events: events,
		Path:   path,
	}, nil
}

// GetEngineInfo returns the ENGINE_INFO event if present.
func (t *Transcript) GetEngineInfo() *TranscriptEvent {
	for i := range t.Events {
		if t.Events[i].Type == EventEngineInfo {
			return &t.Events[i]
		}
	}
	return nil
}

// GetModules returns all discovered modules.
func (t *Transcript) GetModules() []string {
	var modules []string
	seen := make(map[string]bool)
	for _, event := range t.Events {
		if event.Type == EventModuleDiscovered && !seen[event.Module] {
			modules = append(modules, event.Module)
			seen[event.Module] = true
		}
	}
	return modules
}

// GetErrors returns all error events.
func (t *Transcript) GetErrors() []TranscriptEvent {
	var errors []TranscriptEvent
	for _, event := range t.Events {
		if event.Type == EventError {
			errors = append(errors, event)
		}
	}
	return errors
}

// GetWarnings returns all warning events.
func (t *Transcript) GetWarnings() []TranscriptEvent {
	var warnings []TranscriptEvent
	for _, event := range t.Events {
		if event.Type == EventWarn {
			warnings = append(warnings, event)
		}
	}
	return warnings
}

// HasErrors returns true if the transcript contains any error events.
func (t *Transcript) HasErrors() bool {
	for _, event := range t.Events {
		if event.Type == EventError {
			return true
		}
	}
	return false
}

// EventCount returns the total number of events.
func (t *Transcript) EventCount() int {
	return len(t.Events)
}
