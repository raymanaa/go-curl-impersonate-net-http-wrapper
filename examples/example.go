package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	curlhttp "github.com/dstockton/go-curl-impersonate-net-http-wrapper"
)

func main() {
	// Example 1: Basic GET request using the wrapper
	fmt.Println("=== Example 1: GET request with curl-impersonate wrapper ===")

	client := curlhttp.NewClient()
	resp, err := client.Get("https://httpbin.org/get?example=test")
	if err != nil {
		log.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("User-Agent: %s\n", resp.Header.Get("User-Agent"))
	fmt.Printf("Response length: %d bytes\n", len(body))
	fmt.Println()

	// Example 2: POST request with JSON body
	fmt.Println("=== Example 2: POST request with JSON ===")

	jsonData := `{"message": "Hello from curl-impersonate!", "timestamp": "2024"}`
	postResp, err := client.Post("https://httpbin.org/post", "application/json", strings.NewReader(jsonData))
	if err != nil {
		log.Fatalf("POST request failed: %v", err)
	}
	defer postResp.Body.Close()

	postBody, err := io.ReadAll(postResp.Body)
	if err != nil {
		log.Fatalf("Failed to read POST response body: %v", err)
	}

	fmt.Printf("POST Status: %s\n", postResp.Status)
	fmt.Printf("POST Response length: %d bytes\n", len(postBody))
	fmt.Println()

	// Example 3: Compare with standard net/http client
	fmt.Println("=== Example 3: Comparison with standard net/http ===")

	standardClient := &http.Client{}
	standardResp, err := standardClient.Get("https://httpbin.org/get?example=standard")
	if err != nil {
		log.Fatalf("Standard GET request failed: %v", err)
	}
	defer standardResp.Body.Close()

	standardBody, err := io.ReadAll(standardResp.Body)
	if err != nil {
		log.Fatalf("Failed to read standard response body: %v", err)
	}

	fmt.Printf("Standard Status: %s\n", standardResp.Status)
	fmt.Printf("Standard Response length: %d bytes\n", len(standardBody))

	fmt.Println("\n=== Wrapper is working! ===")
	fmt.Println("Both the curl-impersonate wrapper and standard net/http client")
	fmt.Println("can make HTTP requests, with the wrapper providing browser impersonation.")
}
