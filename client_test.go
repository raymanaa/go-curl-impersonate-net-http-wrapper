package curlhttp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

// HttpbinResponse represents the common structure of httpbin.org responses
type HttpbinResponse struct {
	Args    map[string]string `json:"args"`
	Headers map[string]string `json:"headers"`
	Origin  string            `json:"origin"`
	URL     string            `json:"url"`
	Data    string            `json:"data,omitempty"`
	Form    map[string]string `json:"form,omitempty"`
	Files   map[string]string `json:"files,omitempty"`
	JSON    interface{}       `json:"json,omitempty"`
}

// createMockServer creates a test HTTP server for scale and load testing
func createMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/get":
			// Handle GET requests
			w.Header().Set("Content-Type", "application/json")
			id := r.URL.Query().Get("id")
			response := map[string]interface{}{
				"args": r.URL.Query(),
				"headers": func() map[string]string {
					headers := make(map[string]string)
					for k, v := range r.Header {
						if len(v) > 0 {
							headers[k] = v[0]
						}
					}
					return headers
				}(),
				"origin": r.RemoteAddr,
				"url":    fmt.Sprintf("%s%s", "http://"+r.Host, r.RequestURI),
				"method": r.Method,
				"id":     id,
			}
			json.NewEncoder(w).Encode(response)

		case "/post":
			// Handle POST requests
			w.Header().Set("Content-Type", "application/json")
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read body", http.StatusInternalServerError)
				return
			}
			response := map[string]interface{}{
				"args": r.URL.Query(),
				"headers": func() map[string]string {
					headers := make(map[string]string)
					for k, v := range r.Header {
						if len(v) > 0 {
							headers[k] = v[0]
						}
					}
					return headers
				}(),
				"origin": r.RemoteAddr,
				"url":    fmt.Sprintf("%s%s", "http://"+r.Host, r.RequestURI),
				"method": r.Method,
				"data":   string(body),
			}
			json.NewEncoder(w).Encode(response)

		default:
			// Handle other requests
			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"method": r.Method,
				"path":   r.URL.Path,
				"status": "ok",
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
}

// TestGetRequest tests that GET requests work with both clients
func TestGetRequest(t *testing.T) {
	testURL := "https://httpbin.org/get?test=value"

	// Test with standard net/http client
	standardClient := &http.Client{}
	standardResp, err := standardClient.Get(testURL)
	if err != nil {
		t.Skipf("Standard client GET failed (httpbin.org may be unavailable): %v", err)
	}
	defer standardResp.Body.Close()

	// Skip test if httpbin is returning errors
	if standardResp.StatusCode >= 500 {
		t.Skipf("httpbin.org is returning server errors (%d), skipping comparison test", standardResp.StatusCode)
	}

	standardBody, err := io.ReadAll(standardResp.Body)
	if err != nil {
		t.Fatalf("Failed to read standard response body: %v", err)
	}

	// Test with our custom client
	customClient := NewClient()
	customResp, err := customClient.Get(testURL)
	if err != nil {
		t.Fatalf("Custom client GET failed: %v", err)
	}
	defer customResp.Body.Close()

	// Basic functionality test - our client should work even if comparison fails
	if customResp.StatusCode < 200 || customResp.StatusCode >= 400 {
		t.Errorf("Custom client returned error status: %d", customResp.StatusCode)
	}

	customBody, err := io.ReadAll(customResp.Body)
	if err != nil {
		t.Fatalf("Failed to read custom response body: %v", err)
	}

	// Only compare if both requests succeeded
	if standardResp.StatusCode == 200 && customResp.StatusCode == 200 {
		// Parse and compare JSON responses (excluding dynamic fields)
		var standardData, customData HttpbinResponse
		if err := json.Unmarshal(standardBody, &standardData); err != nil {
			t.Logf("Failed to parse standard response JSON (may be HTML error page): %v", err)
			return
		}
		if err := json.Unmarshal(customBody, &customData); err != nil {
			t.Logf("Failed to parse custom response JSON (may be HTML error page): %v", err)
			return
		}

		// Compare static fields
		compareHttpbinResponses(t, "GET", standardData, customData)
	}
}

// TestPostRequest tests that POST requests work with both clients
func TestPostRequest(t *testing.T) {
	testURL := "https://httpbin.org/post"
	testData := `{"test": "data", "number": 42}`

	// Test with standard net/http client
	standardClient := &http.Client{}
	standardResp, err := standardClient.Post(testURL, "application/json", strings.NewReader(testData))
	if err != nil {
		t.Skipf("Standard client POST failed (httpbin.org may be unavailable): %v", err)
	}
	defer standardResp.Body.Close()

	// Skip test if httpbin is returning errors
	if standardResp.StatusCode >= 500 {
		t.Skipf("httpbin.org is returning server errors (%d), skipping comparison test", standardResp.StatusCode)
	}

	standardBody, err := io.ReadAll(standardResp.Body)
	if err != nil {
		t.Fatalf("Failed to read standard response body: %v", err)
	}

	// Test with our custom client
	customClient := NewClient()
	customResp, err := customClient.Post(testURL, "application/json", strings.NewReader(testData))
	if err != nil {
		t.Fatalf("Custom client POST failed: %v", err)
	}
	defer customResp.Body.Close()

	// Basic functionality test - our client should work even if comparison fails
	if customResp.StatusCode < 200 || customResp.StatusCode >= 400 {
		t.Errorf("Custom client returned error status: %d", customResp.StatusCode)
	}

	customBody, err := io.ReadAll(customResp.Body)
	if err != nil {
		t.Fatalf("Failed to read custom response body: %v", err)
	}

	// Only compare if both requests succeeded
	if standardResp.StatusCode == 200 && customResp.StatusCode == 200 {
		// Parse and compare JSON responses
		var standardData, customData HttpbinResponse
		if err := json.Unmarshal(standardBody, &standardData); err != nil {
			t.Logf("Failed to parse standard response JSON (may be HTML error page): %v", err)
			return
		}
		if err := json.Unmarshal(customBody, &customData); err != nil {
			t.Logf("Failed to parse custom response JSON (may be HTML error page): %v", err)
			return
		}

		// Compare static fields
		compareHttpbinResponses(t, "POST", standardData, customData)

		// Verify POST data was received correctly
		if standardData.Data != customData.Data {
			t.Errorf("POST data differs: standard=%s, custom=%s", standardData.Data, customData.Data)
		}
	}
}

// Test using Do method directly
func TestDoMethod(t *testing.T) {
	testURL := "https://httpbin.org/get"

	// Create request
	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Test with standard client
	standardClient := &http.Client{}
	standardResp, err := standardClient.Do(req)
	if err != nil {
		t.Fatalf("Standard client Do failed: %v", err)
	}
	defer standardResp.Body.Close()

	// Test with custom client (need to recreate request because body may be consumed)
	req2, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		t.Fatalf("Failed to create second request: %v", err)
	}

	customClient := NewClient()
	customResp, err := customClient.Do(req2)
	if err != nil {
		t.Fatalf("Custom client Do failed: %v", err)
	}
	defer customResp.Body.Close()

	// Compare status codes
	if standardResp.StatusCode != customResp.StatusCode {
		t.Errorf("Status codes differ: standard=%d, custom=%d", standardResp.StatusCode, customResp.StatusCode)
	}
}

// TestCustomHeaders tests that custom headers are properly sent
func TestCustomHeaders(t *testing.T) {
	testURL := "https://httpbin.org/headers"

	// Create request with custom headers
	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("X-Test-Header", "test-value")
	req.Header.Set("User-Agent", "test-agent/1.0")

	// Test with custom client
	customClient := NewClient()
	customResp, err := customClient.Do(req)
	if err != nil {
		t.Fatalf("Custom client request failed: %v", err)
	}
	defer customResp.Body.Close()

	// Skip test if httpbin is returning errors
	if customResp.StatusCode >= 500 {
		t.Skipf("httpbin.org is returning server errors (%d), skipping test", customResp.StatusCode)
	}

	body, err := io.ReadAll(customResp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Basic functionality test
	if customResp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", customResp.StatusCode)
		return
	}

	var response HttpbinResponse
	if err := json.Unmarshal(body, &response); err != nil {
		t.Logf("Failed to parse response JSON (may be HTML error page): %v", err)
		return
	}

	// Verify custom headers were sent
	if response.Headers["X-Test-Header"] != "test-value" {
		t.Errorf("Custom header not found or incorrect: got %s, want test-value", response.Headers["X-Test-Header"])
	}
}

// compareHttpbinResponses compares two httpbin responses, ignoring dynamic fields
func compareHttpbinResponses(t *testing.T, method string, standard, custom HttpbinResponse) {
	// Compare args (query parameters)
	if !reflect.DeepEqual(standard.Args, custom.Args) {
		t.Errorf("%s args differ: standard=%v, custom=%v", method, standard.Args, custom.Args)
	}

	// Compare URL (should be exactly the same)
	if standard.URL != custom.URL {
		t.Errorf("%s URL differs: standard=%s, custom=%s", method, standard.URL, custom.URL)
	}

	// Compare some static headers (ignoring dynamic ones)
	staticHeaders := []string{"Content-Type", "Content-Length"}
	for _, header := range staticHeaders {
		if standard.Headers[header] != custom.Headers[header] {
			// Only error if both have the header but values differ
			if standard.Headers[header] != "" && custom.Headers[header] != "" {
				t.Errorf("%s header %s differs: standard=%s, custom=%s",
					method, header, standard.Headers[header], custom.Headers[header])
			}
		}
	}

	// Note: We intentionally skip comparing dynamic headers like:
	// - Date
	// - User-Agent (may differ due to impersonation)
	// - X-Amzn-Trace-Id
	// - And other server-generated or client-specific headers
}

// Benchmark comparing performance of both clients
func BenchmarkGetPerformance(b *testing.B) {
	testURL := "https://httpbin.org/get"

	b.Run("NetHTTP", func(b *testing.B) {
		client := &http.Client{}
		for i := 0; i < b.N; i++ {
			resp, err := client.Get(testURL)
			if err != nil {
				b.Fatalf("Request failed: %v", err)
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})

	b.Run("CurlImpersonate", func(b *testing.B) {
		client := NewClient()
		for i := 0; i < b.N; i++ {
			resp, err := client.Get(testURL)
			if err != nil {
				b.Fatalf("Request failed: %v", err)
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})
}
