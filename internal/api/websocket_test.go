package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()
	if hub == nil {
		t.Fatal("NewHub returned nil")
	}
	if hub.clients == nil {
		t.Error("Hub clients map is nil")
	}
	if hub.broadcast == nil {
		t.Error("Hub broadcast channel is nil")
	}
	if hub.register == nil {
		t.Error("Hub register channel is nil")
	}
	if hub.unregister == nil {
		t.Error("Hub unregister channel is nil")
	}
}

func TestHubRunAndBroadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create a test server with WebSocket handler
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("Failed to upgrade: %v", err)
			return
		}

		client := &Client{
			hub:  hub,
			conn: conn,
			send: make(chan []byte, 256),
		}

		hub.register <- client
		go client.writePump()
		go client.readPump()
	}))
	defer server.Close()

	// Connect WebSocket client
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Wait for client to register
	time.Sleep(100 * time.Millisecond)

	// Broadcast a message
	testMsg := ProgressMessage{
		Type:      "progress",
		Operation: "test",
		Stage:     "testing",
		Progress:  50,
		Message:   "Test message",
	}
	hub.Broadcast(testMsg)

	// Read the message
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	var received ProgressMessage
	if err := json.Unmarshal(data, &received); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if received.Type != testMsg.Type {
		t.Errorf("Expected type %s, got %s", testMsg.Type, received.Type)
	}
	if received.Operation != testMsg.Operation {
		t.Errorf("Expected operation %s, got %s", testMsg.Operation, received.Operation)
	}
	if received.Stage != testMsg.Stage {
		t.Errorf("Expected stage %s, got %s", testMsg.Stage, received.Stage)
	}
	if received.Progress != testMsg.Progress {
		t.Errorf("Expected progress %d, got %d", testMsg.Progress, received.Progress)
	}
	if received.Message != testMsg.Message {
		t.Errorf("Expected message %s, got %s", testMsg.Message, received.Message)
	}
	if received.Timestamp == "" {
		t.Error("Timestamp should be automatically set")
	}
}

func TestBroadcastHelpers(t *testing.T) {
	// Save original hub and restore after test
	originalHub := GlobalHub
	defer func() { GlobalHub = originalHub }()

	hub := NewHub()
	GlobalHub = hub
	go hub.Run()

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		client := &Client{
			hub:  hub,
			conn: conn,
			send: make(chan []byte, 256),
		}

		hub.register <- client
		go client.writePump()
		go client.readPump()
	}))
	defer server.Close()

	// Connect client
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	time.Sleep(100 * time.Millisecond)

	// Test BroadcastProgress
	BroadcastProgress("convert", "extracting", "Extracting IR", 25)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read progress: %v", err)
	}

	var progress ProgressMessage
	if err := json.Unmarshal(data, &progress); err != nil {
		t.Fatalf("Failed to unmarshal progress: %v", err)
	}
	if progress.Type != "progress" {
		t.Errorf("Expected type 'progress', got %s", progress.Type)
	}
	if progress.Progress != 25 {
		t.Errorf("Expected progress 25, got %d", progress.Progress)
	}

	// Test BroadcastComplete
	BroadcastComplete("convert", "Conversion completed", map[string]interface{}{
		"output": "test.capsule",
	})
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, data, err = conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read complete: %v", err)
	}

	var complete ProgressMessage
	if err := json.Unmarshal(data, &complete); err != nil {
		t.Fatalf("Failed to unmarshal complete: %v", err)
	}
	if complete.Type != "complete" {
		t.Errorf("Expected type 'complete', got %s", complete.Type)
	}
	if complete.Progress != 100 {
		t.Errorf("Expected progress 100, got %d", complete.Progress)
	}
	if complete.Data == nil {
		t.Error("Expected data map to be set")
	}

	// Test BroadcastError
	BroadcastError("convert", "Conversion failed")
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, data, err = conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read error: %v", err)
	}

	var errorMsg ProgressMessage
	if err := json.Unmarshal(data, &errorMsg); err != nil {
		t.Fatalf("Failed to unmarshal error: %v", err)
	}
	if errorMsg.Type != "error" {
		t.Errorf("Expected type 'error', got %s", errorMsg.Type)
	}
}

func TestHandleWebSocket(t *testing.T) {
	// Save original hub and restore after test
	originalHub := GlobalHub
	defer func() { GlobalHub = originalHub }()

	hub := NewHub()
	GlobalHub = hub
	go hub.Run()

	server := httptest.NewServer(http.HandlerFunc(handleWebSocket))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Errorf("Expected status 101, got %d", resp.StatusCode)
	}

	// Verify client was registered
	time.Sleep(100 * time.Millisecond)
	hub.mu.RLock()
	clientCount := len(hub.clients)
	hub.mu.RUnlock()

	if clientCount != 1 {
		t.Errorf("Expected 1 client, got %d", clientCount)
	}
}

func TestHandleWebSocketNoHub(t *testing.T) {
	// Save original hub and restore after test
	originalHub := GlobalHub
	GlobalHub = nil
	defer func() { GlobalHub = originalHub }()

	req := httptest.NewRequest("GET", "/ws", nil)
	w := httptest.NewRecorder()

	handleWebSocket(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestMultipleClients(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		client := &Client{
			hub:  hub,
			conn: conn,
			send: make(chan []byte, 256),
		}

		hub.register <- client
		go client.writePump()
		go client.readPump()
	}))
	defer server.Close()

	// Connect multiple clients
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect client 1: %v", err)
	}
	defer conn1.Close()

	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect client 2: %v", err)
	}
	defer conn2.Close()

	time.Sleep(100 * time.Millisecond)

	// Broadcast message
	testMsg := ProgressMessage{
		Type:      "progress",
		Operation: "test",
		Progress:  75,
		Message:   "Multi-client test",
	}
	hub.Broadcast(testMsg)

	// Both clients should receive the message
	for i, conn := range []*websocket.Conn{conn1, conn2} {
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, data, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("Client %d failed to read: %v", i+1, err)
		}

		var received ProgressMessage
		if err := json.Unmarshal(data, &received); err != nil {
			t.Fatalf("Client %d failed to unmarshal: %v", i+1, err)
		}

		if received.Progress != testMsg.Progress {
			t.Errorf("Client %d: expected progress %d, got %d", i+1, testMsg.Progress, received.Progress)
		}
	}
}

func TestClientDisconnect(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		client := &Client{
			hub:  hub,
			conn: conn,
			send: make(chan []byte, 256),
		}

		hub.register <- client
		go client.writePump()
		go client.readPump()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify client is registered
	hub.mu.RLock()
	clientCount := len(hub.clients)
	hub.mu.RUnlock()
	if clientCount != 1 {
		t.Errorf("Expected 1 client before disconnect, got %d", clientCount)
	}

	// Disconnect
	conn.Close()
	time.Sleep(100 * time.Millisecond)

	// Verify client is unregistered
	hub.mu.RLock()
	clientCount = len(hub.clients)
	hub.mu.RUnlock()
	if clientCount != 0 {
		t.Errorf("Expected 0 clients after disconnect, got %d", clientCount)
	}
}
