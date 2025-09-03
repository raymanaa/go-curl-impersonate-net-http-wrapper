package curlhttp

import (
	"net/http"
	"testing"
	"time"
)

// TestNewClient tests client creation
func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
	if client.Transport == nil {
		t.Fatal("NewClient() created client with nil Transport")
	}
	if client.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", client.Timeout)
	}
}

// TestNewClientWithTarget tests client creation with custom target
func TestNewClientWithTarget(t *testing.T) {
	target := "firefox102"
	client := NewClientWithTarget(target)
	if client == nil {
		t.Fatal("NewClientWithTarget() returned nil")
	}

	transport, ok := client.Transport.(*Transport)
	if !ok {
		t.Fatal("Client transport is not *Transport")
	}

	if transport.ImpersonateTarget != target {
		t.Errorf("Expected target %s, got %s", target, transport.ImpersonateTarget)
	}
	if !transport.UseDefaultHeaders {
		t.Error("Expected UseDefaultHeaders to be true")
	}
}

// TestNewTransport tests transport creation
func TestNewTransport(t *testing.T) {
	transport := NewTransport()
	if transport == nil {
		t.Fatal("NewTransport() returned nil")
	}
	if transport.ImpersonateTarget != "chrome136" {
		t.Errorf("Expected default target chrome136, got %s", transport.ImpersonateTarget)
	}
	if !transport.UseDefaultHeaders {
		t.Error("Expected UseDefaultHeaders to be true")
	}
}

// TestTransportRoundTripNilRequest tests error handling for nil request
func TestTransportRoundTripNilRequest(t *testing.T) {
	transport := NewTransport()
	_, err := transport.RoundTrip(nil)
	if err == nil {
		t.Error("Expected error for nil request, got nil")
	}
	if err.Error() != "request cannot be nil" {
		t.Errorf("Expected 'request cannot be nil' error, got: %v", err)
	}
}

// TestTransportRoundTripNilURL tests error handling for nil URL
func TestTransportRoundTripNilURL(t *testing.T) {
	transport := NewTransport()
	req := &http.Request{} // Request with nil URL
	_, err := transport.RoundTrip(req)
	if err == nil {
		t.Error("Expected error for nil URL, got nil")
	}
	if err.Error() != "request URL cannot be nil" {
		t.Errorf("Expected 'request URL cannot be nil' error, got: %v", err)
	}
}

// TestDefaultClient tests the default client
func TestDefaultClient(t *testing.T) {
	if DefaultClient == nil {
		t.Fatal("DefaultClient is nil")
	}
	if DefaultClient.Transport == nil {
		t.Fatal("DefaultClient.Transport is nil")
	}
	if DefaultClient.Timeout != 30*time.Second {
		t.Errorf("Expected DefaultClient timeout 30s, got %v", DefaultClient.Timeout)
	}

	transport, ok := DefaultClient.Transport.(*Transport)
	if !ok {
		t.Fatal("DefaultClient transport is not *Transport")
	}

	if transport.ImpersonateTarget != "chrome136" {
		t.Errorf("Expected DefaultClient target chrome136, got %s", transport.ImpersonateTarget)
	}
}

// TestParseHeaders tests header parsing functionality
func TestParseHeaders(t *testing.T) {
	headerData := "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: 123\r\nX-Custom: test-value\r\n"

	headers := parseHeaders(headerData)

	// Check that headers were parsed correctly
	if headers.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type: application/json, got: %s", headers.Get("Content-Type"))
	}

	if headers.Get("Content-Length") != "123" {
		t.Errorf("Expected Content-Length: 123, got: %s", headers.Get("Content-Length"))
	}

	if headers.Get("X-Custom") != "test-value" {
		t.Errorf("Expected X-Custom: test-value, got: %s", headers.Get("X-Custom"))
	}

	// Check that status line was ignored
	if headers.Get("HTTP/1.1") != "" {
		t.Error("Status line should not be parsed as a header")
	}
}

// TestParseHeadersEmptyInput tests header parsing with empty input
func TestParseHeadersEmptyInput(t *testing.T) {
	headers := parseHeaders("")
	if len(headers) != 0 {
		t.Errorf("Expected empty headers map, got %d headers", len(headers))
	}
}

// TestParseHeadersWithNewlines tests header parsing with different newline styles
func TestParseHeadersWithNewlines(t *testing.T) {
	headerData := "HTTP/1.1 200 OK\nContent-Type: application/json\nContent-Length: 123\n"

	headers := parseHeaders(headerData)

	if headers.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type: application/json, got: %s", headers.Get("Content-Type"))
	}
}
