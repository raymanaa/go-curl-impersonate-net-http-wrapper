package main

import (
	"fmt"
	"io"
	"log"

	"net/http"

	curlhttp "github.com/dstockton/go-curl-impersonate-net-http-wrapper"
)

func main() {
	getResponse(http.Get)
	getResponse(curlhttp.Get)
}

func getResponse(lib func(string) (*http.Response, error)) {
	fmt.Println("🌐 Making HTTP request...")

	fmt.Printf("Using library: %T\n", lib)

	resp, err := lib("https://httpbin.org/user-agent")
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Response: %s\n", string(body))
	fmt.Println("✅ Request completed")
}
