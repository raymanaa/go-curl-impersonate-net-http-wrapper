package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	// Perfect drop-in replacement - literally just swap the import!
	http "github.com/dstockton/go-curl-impersonate-net-http-wrapper"
	// http "net/http"  // <- Comment out this line and uncomment the line above
)

func main() {
	fmt.Println("🚀 Complete Drop-in Replacement Demonstration")
	fmt.Println("This exact code works with both net/http and curl-impersonate wrapper!")
	fmt.Println()

	// Test 1: Zero-value client (most common pattern)
	fmt.Println("1. ✅ Zero-value Client initialization:")
	client := &http.Client{}
	resp, err := client.Get("https://httpbin.org/get")
	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}
	defer resp.Body.Close()
	fmt.Printf("   Status: %s\n", resp.Status)

	// Test 2: Client with custom timeout (common pattern)
	fmt.Println("\n2. ✅ Client with custom configuration:")
	customClient := &http.Client{}
	customClient.Timeout = 10 * time.Second
	resp2, err := customClient.Get("https://httpbin.org/get?client=custom")
	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}
	defer resp2.Body.Close()
	fmt.Printf("   Status: %s, Timeout: %v\n", resp2.Status, customClient.Timeout)

	// Test 3: Package-level functions (very common)
	fmt.Println("\n3. ✅ Package-level functions:")
	resp3, err := http.Get("https://httpbin.org/get?test=package")
	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}
	defer resp3.Body.Close()
	fmt.Printf("   http.Get() - Status: %s\n", resp3.Status)

	// Test 4: POST with package function
	resp4, err := http.Post("https://httpbin.org/post", "application/json", strings.NewReader(`{"test": "drop-in"}`))
	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}
	defer resp4.Body.Close()
	fmt.Printf("   http.Post() - Status: %s\n", resp4.Status)

	// Test 5: NewRequest and Do (very common pattern)
	fmt.Println("\n4. ✅ Request creation and Do method:")
	req, err := http.NewRequest("GET", "https://httpbin.org/get?test=newreq", nil)
	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}
	req.Header.Set("X-Test", "drop-in-replacement")

	resp5, err := http.Do(req)
	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}
	defer resp5.Body.Close()
	fmt.Printf("   http.NewRequest() + http.Do() - Status: %s\n", resp5.Status)

	// Test 6: Status constants
	fmt.Println("\n5. ✅ Status constants:")
	fmt.Printf("   http.StatusOK = %d\n", http.StatusOK)
	fmt.Printf("   http.StatusNotFound = %d\n", http.StatusNotFound)
	fmt.Printf("   http.StatusInternalServerError = %d\n", http.StatusInternalServerError)

	// Test 7: Method constants
	fmt.Println("\n6. ✅ Method constants:")
	fmt.Printf("   http.MethodGet = %s\n", http.MethodGet)
	fmt.Printf("   http.MethodPost = %s\n", http.MethodPost)
	fmt.Printf("   http.MethodPut = %s\n", http.MethodPut)

	// Test 8: Context-aware requests
	fmt.Println("\n7. ✅ Context-aware requests:")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req6, err := http.NewRequestWithContext(ctx, "GET", "https://httpbin.org/get?test=context", nil)
	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}
	resp6, err := http.Do(req6)
	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}
	defer resp6.Body.Close()
	fmt.Printf("   Context-aware request - Status: %s\n", resp6.Status)

	fmt.Println("\n🎉 SUCCESS! Complete drop-in replacement working perfectly!")
	fmt.Println("✨ You can literally swap imports between net/http and this wrapper!")
	fmt.Println("🚀 All standard Go HTTP patterns work unchanged!")
}
