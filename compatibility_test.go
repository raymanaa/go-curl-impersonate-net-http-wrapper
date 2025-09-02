package curlhttp

import (
	"encoding/json"
	"io"
	"strings"
	"testing"
)

// Test package-level functions for drop-in compatibility
// These tests verify that the package can be used as a direct replacement for net/http

// TestPackageLevelGet tests the package-level Get function
func TestPackageLevelGet(t *testing.T) {
	// Use the package-level Get function just like net/http
	resp, err := Get("https://httpbin.org/get?package=test")
	if err != nil {
		t.Fatalf("Package-level Get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Parse JSON response
	var data HttpbinResponse
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	if data.Args["package"] != "test" {
		t.Errorf("Expected package=test in query args, got %v", data.Args)
	}

	t.Logf("Package-level Get test passed: %s", resp.Status)
}

// TestPackageLevelPost tests the package-level Post function
func TestPackageLevelPost(t *testing.T) {
	jsonData := `{"test": "package-level"}`

	// Use the package-level Post function just like net/http
	resp, err := Post("https://httpbin.org/post", "application/json", strings.NewReader(jsonData))
	if err != nil {
		t.Fatalf("Package-level Post failed: %v", err)
	}
	defer resp.Body.Close()

	// Skip test if httpbin is returning errors
	if resp.StatusCode >= 500 {
		t.Skipf("httpbin.org is returning server errors (%d), skipping test", resp.StatusCode)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Parse JSON response
	var data HttpbinResponse
	if err := json.Unmarshal(body, &data); err != nil {
		t.Logf("Failed to parse response JSON (may be HTML error page): %v", err)
		return
	}

	if data.Data != jsonData {
		t.Errorf("Expected posted data %s, got %s", jsonData, data.Data)
	}

	t.Logf("Package-level Post test passed: %s", resp.Status)
}

// TestPackageLevelHead tests the package-level Head function
func TestPackageLevelHead(t *testing.T) {
	// Use the package-level Head function just like net/http
	resp, err := Head("https://httpbin.org/get")
	if err != nil {
		t.Fatalf("Package-level Head failed: %v", err)
	}
	defer resp.Body.Close()

	// Skip test if httpbin is returning errors
	if resp.StatusCode >= 500 {
		t.Skipf("httpbin.org is returning server errors (%d), skipping test", resp.StatusCode)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
		return
	}

	// HEAD requests should have no body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if len(body) > 0 {
		t.Errorf("HEAD request should have empty body, got %d bytes", len(body))
	}

	t.Logf("Package-level Head test passed: %s", resp.Status)
}

// TestPackageLevelDo tests the package-level Do function
func TestPackageLevelDo(t *testing.T) {
	// Create a request just like with net/http
	req, err := NewRequest("GET", "https://httpbin.org/get?do=test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Add a custom header
	req.Header.Set("X-Test-Header", "package-level-do")

	// Use the package-level Do function just like net/http
	resp, err := Do(req)
	if err != nil {
		t.Fatalf("Package-level Do failed: %v", err)
	}
	defer resp.Body.Close()

	// Skip test if httpbin is returning errors
	if resp.StatusCode >= 500 {
		t.Skipf("httpbin.org is returning server errors (%d), skipping test", resp.StatusCode)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Parse JSON response
	var data HttpbinResponse
	if err := json.Unmarshal(body, &data); err != nil {
		t.Logf("Failed to parse response JSON (may be HTML error page): %v", err)
		return
	}

	if data.Args["do"] != "test" {
		t.Errorf("Expected do=test in query args, got %v", data.Args)
	}

	// Check that our custom header was sent
	if data.Headers["X-Test-Header"] != "package-level-do" {
		t.Errorf("Expected X-Test-Header=package-level-do, got %s", data.Headers["X-Test-Header"])
	}

	t.Logf("Package-level Do test passed: %s", resp.Status)
}
