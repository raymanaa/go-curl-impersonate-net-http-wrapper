package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	// Just swap this import line to enable browser impersonation!
	// import "net/http"  // <- Standard library
	http "github.com/dstockton/go-curl-impersonate-net-http-wrapper" // <- With browser impersonation
)

func main() {
	fmt.Println("=== Import Swap Example ===")
	fmt.Println("This code works identically whether you import net/http or the wrapper!")
	fmt.Println()

	// Example 1: Package-level Get function (exactly like net/http)
	fmt.Println("1. Making GET request...")
	resp, err := http.Get("https://httpbin.org/get?source=import-swap")
	if err != nil {
		log.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	fmt.Printf("   Status: %s\n", resp.Status)
	fmt.Printf("   Content-Length: %d bytes\n", len(body))
	fmt.Println()

	// Example 2: Package-level Post function (exactly like net/http)
	fmt.Println("2. Making POST request...")
	jsonData := `{"message": "Hello from import-swap example", "browser": "chrome136"}`
	resp2, err := http.Post("https://httpbin.org/post", "application/json", strings.NewReader(jsonData))
	if err != nil {
		log.Fatalf("POST request failed: %v", err)
	}
	defer resp2.Body.Close()

	body2, err := io.ReadAll(resp2.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	fmt.Printf("   Status: %s\n", resp2.Status)
	fmt.Printf("   Content-Length: %d bytes\n", len(body2))
	fmt.Println()

	// Example 3: Creating custom request (exactly like net/http)
	fmt.Println("3. Making custom request with headers...")
	req, err := http.NewRequest("GET", "https://httpbin.org/headers", nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	// Add custom headers
	req.Header.Set("X-Custom-Header", "import-swap-test")
	req.Header.Set("X-Browser-Impersonate", "true")

	resp3, err := http.Do(req)
	if err != nil {
		log.Fatalf("Custom request failed: %v", err)
	}
	defer resp3.Body.Close()

	body3, err := io.ReadAll(resp3.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	fmt.Printf("   Status: %s\n", resp3.Status)
	fmt.Printf("   Content-Length: %d bytes\n", len(body3))
	fmt.Println()

	fmt.Println("=== All requests completed successfully! ===")
	fmt.Println("The browser impersonation is working transparently.")
}
