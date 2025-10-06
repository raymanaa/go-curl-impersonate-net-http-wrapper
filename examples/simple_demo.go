package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	// This is how you use the wrapper - just import it like net/http!
	http "github.com/dstockton/go-curl-impersonate-net-http-wrapper"
	// http "net/http"  // <- You can literally swap this line!
)

func main() {
	fmt.Println("🚀 Simple Getting Started Example")
	fmt.Println("This wrapper is a drop-in replacement for net/http with browser impersonation!")
	fmt.Println()

	// Example 1: Simple GET request (just like net/http!)
	fmt.Println("1. 📥 Basic GET request:")
	resp, err := http.Get("https://httpbin.org/get?demo=simple")
	if err != nil {
		log.Fatalf("❌ GET failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("❌ Failed to read response: %v", err)
	}
	fmt.Printf("   ✅ Status: %s, Length: %d bytes\n", resp.Status, len(body))

	// Example 2: POST with JSON (exactly like net/http!)
	fmt.Println("2. 📤 POST with JSON:")
	jsonData := `{"message": "Hello from browser!", "demo": "simple"}`
	postResp, err := http.Post("https://httpbin.org/post", "application/json", strings.NewReader(jsonData))
	if err != nil {
		log.Fatalf("❌ POST failed: %v", err)
	}
	defer postResp.Body.Close()

	postBody, err := io.ReadAll(postResp.Body)
	if err != nil {
		log.Fatalf("❌ Failed to read POST response: %v", err)
	}
	fmt.Printf("   ✅ Status: %s, Length: %d bytes\n", postResp.Status, len(postBody))

	// Example 3: Using a Client instance (exactly like net/http!)
	fmt.Println("3. 🛠️ Custom client:")
	client := &http.Client{} // Zero-value client works just like net/http!
	clientResp, err := client.Get("https://httpbin.org/get?demo=client")
	if err != nil {
		log.Fatalf("❌ Client GET failed: %v", err)
	}
	defer clientResp.Body.Close()
	fmt.Printf("   ✅ Status: %s\n", clientResp.Status)

	fmt.Println()
	fmt.Println("🎉 That's it! The wrapper works exactly like net/http!")
	fmt.Println("🕵️ But now your requests look like they're coming from a real browser!")
	fmt.Println("🔄 Try swapping the import to see it works with standard net/http too!")
}
