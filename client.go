// Package curlhttp provides a drop-in replacement for net/http that uses
// curl-impersonate for browser impersonation to avoid detection by websites
// that block automated requests.
//
// This package re-exports all net/http types, constants, and functions while
// using curl-impersonate as the underlying HTTP client. Simply replace your
// net/http import with this package for seamless browser impersonation.
//
// Example usage:
//
//	// Instead of: import "net/http"
//	import http "github.com/dstockton/go-curl-impersonate-net-http-wrapper"
//
//	// All your existing net/http code works unchanged!
//	resp, err := http.Get("https://example.com")
package curlhttp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	curl "github.com/BridgeSenseDev/go-curl-impersonate"
)

// Re-export all net/http types for drop-in compatibility
type (
	Request        = http.Request
	Response       = http.Response
	Header         = http.Header
	Cookie         = http.Cookie
	RoundTripper   = http.RoundTripper
	CookieJar      = http.CookieJar
	File           = http.File
	FileSystem     = http.FileSystem
	Flusher        = http.Flusher
	Hijacker       = http.Hijacker
	Handler        = http.Handler
	HandlerFunc    = http.HandlerFunc
	ServeMux       = http.ServeMux
	Server         = http.Server
	ResponseWriter = http.ResponseWriter
)

// Re-export common constants
const (
	MethodGet     = http.MethodGet
	MethodHead    = http.MethodHead
	MethodPost    = http.MethodPost
	MethodPut     = http.MethodPut
	MethodPatch   = http.MethodPatch
	MethodDelete  = http.MethodDelete
	MethodConnect = http.MethodConnect
	MethodOptions = http.MethodOptions
	MethodTrace   = http.MethodTrace
)

// Re-export status codes
const (
	StatusOK                  = http.StatusOK
	StatusCreated             = http.StatusCreated
	StatusAccepted            = http.StatusAccepted
	StatusNoContent           = http.StatusNoContent
	StatusBadRequest          = http.StatusBadRequest
	StatusUnauthorized        = http.StatusUnauthorized
	StatusForbidden           = http.StatusForbidden
	StatusNotFound            = http.StatusNotFound
	StatusInternalServerError = http.StatusInternalServerError
	// Add more as needed
)

// Re-export common functions
var (
	NewRequest            = http.NewRequest
	NewRequestWithContext = http.NewRequestWithContext
	ReadResponse          = http.ReadResponse
	ParseHTTPVersion      = http.ParseHTTPVersion
	ParseTime             = http.ParseTime
	StatusText            = http.StatusText
	CanonicalHeaderKey    = http.CanonicalHeaderKey
	DetectContentType     = http.DetectContentType
	Error                 = http.Error
	NotFound              = http.NotFound
	Redirect              = http.Redirect
	Serve                 = http.Serve
	ServeFile             = http.ServeFile
	StripPrefix           = http.StripPrefix
	TimeoutHandler        = http.TimeoutHandler
	NewServeMux           = http.NewServeMux
	HandleFunc            = http.HandleFunc
	Handle                = http.Handle
	ListenAndServe        = http.ListenAndServe
	ListenAndServeTLS     = http.ListenAndServeTLS
)

var globalInitOnce sync.Once

// initCurl ensures curl is globally initialized
func initCurl() {
	globalInitOnce.Do(func() {
		curl.GlobalInit(curl.GLOBAL_ALL)
	})
}

// writeData is the callback function for writing response data.
// It safely writes data to the provided io.Writer and returns false on error.
func writeData(ptr []byte, userdata interface{}) bool {
	writer, ok := userdata.(io.Writer)
	if !ok {
		return false
	}
	_, err := writer.Write(ptr)
	return err == nil
}

// performRequest performs a single HTTP request using curl with temporary files
func performRequest(url, method string, headers map[string]string, body []byte) (*http.Response, error) {
	initCurl()

	easy := curl.EasyInit()
	if easy == nil {
		return nil, fmt.Errorf("failed to initialize curl handle")
	}
	defer easy.Cleanup()

	// Set the URL
	if err := easy.Setopt(curl.OPT_URL, url); err != nil {
		return nil, fmt.Errorf("failed to set URL: %w", err)
	}

	// Set up browser impersonation
	if err := easy.Impersonate("chrome136", true); err != nil {
		return nil, fmt.Errorf("failed to impersonate: %w", err)
	}

	// Disable progress reporting
	if err := easy.Setopt(curl.OPT_NOPROGRESS, true); err != nil {
		return nil, fmt.Errorf("failed to disable progress: %w", err)
	}

	// Set HTTP method
	switch method {
	case "GET":
		// GET is the default
	case "HEAD":
		if err := easy.Setopt(curl.OPT_NOBODY, true); err != nil {
			return nil, fmt.Errorf("failed to set HEAD method: %w", err)
		}
	case "POST":
		if err := easy.Setopt(curl.OPT_POST, true); err != nil {
			return nil, fmt.Errorf("failed to set POST method: %w", err)
		}
		if len(body) > 0 {
			if err := easy.Setopt(curl.OPT_POSTFIELDS, body); err != nil {
				return nil, fmt.Errorf("failed to set request body: %w", err)
			}
		}
	case "PUT":
		if err := easy.Setopt(curl.OPT_UPLOAD, true); err != nil {
			return nil, fmt.Errorf("failed to set PUT method: %w", err)
		}
	case "DELETE":
		if err := easy.Setopt(curl.OPT_CUSTOMREQUEST, "DELETE"); err != nil {
			return nil, fmt.Errorf("failed to set DELETE method: %w", err)
		}
	default:
		if err := easy.Setopt(curl.OPT_CUSTOMREQUEST, method); err != nil {
			return nil, fmt.Errorf("failed to set custom method %s: %w", method, err)
		}
	}

	// Set headers
	if len(headers) > 0 {
		var requestHeaders []string
		for name, value := range headers {
			requestHeaders = append(requestHeaders, fmt.Sprintf("%s: %s", name, value))
		}
		if err := easy.Setopt(curl.OPT_HTTPHEADER, requestHeaders); err != nil {
			return nil, fmt.Errorf("failed to set headers: %w", err)
		}
	}

	// Include headers in the response body
	if err := easy.Setopt(curl.OPT_HEADER, true); err != nil {
		return nil, fmt.Errorf("failed to include headers: %w", err)
	}

	// Create temporary file for capturing response (headers + body)
	responseFile, err := os.CreateTemp("", "curl_response_*.tmp")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file for response: %w", err)
	}
	defer os.Remove(responseFile.Name())
	defer responseFile.Close()

	// Set response callback function with file as userdata
	if err := easy.Setopt(curl.OPT_WRITEFUNCTION, writeData); err != nil {
		return nil, fmt.Errorf("failed to set write function: %w", err)
	}
	if err := easy.Setopt(curl.OPT_WRITEDATA, responseFile); err != nil {
		return nil, fmt.Errorf("failed to set write data: %w", err)
	}

	// Perform the request
	if err := easy.Perform(); err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Get response code
	responseCodeInfo, err := easy.Getinfo(uint32(curl.INFO_HTTP_CODE))
	if err != nil {
		return nil, fmt.Errorf("failed to get response code: %w", err)
	}
	responseCode := int(responseCodeInfo.(int64))

	// Read the combined response (headers + body) from temp file
	if _, err := responseFile.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to seek response file: %w", err)
	}
	combinedData, err := io.ReadAll(responseFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Split headers and body - they are separated by double CRLF
	combinedStr := string(combinedData)
	parts := strings.SplitN(combinedStr, "\r\n\r\n", 2)
	if len(parts) < 2 {
		// Try with just \n\n as fallback
		parts = strings.SplitN(combinedStr, "\n\n", 2)
		if len(parts) < 2 {
			return nil, fmt.Errorf("failed to separate headers and body in response")
		}
	}

	headerStr := parts[0]
	bodyStr := parts[1]

	// Parse headers
	responseHeaders := parseHeaders(headerStr)
	responseBodyData := []byte(bodyStr)

	// Create http.Response
	resp := &http.Response{
		Status:        fmt.Sprintf("%d %s", responseCode, http.StatusText(responseCode)),
		StatusCode:    responseCode,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        responseHeaders,
		Body:          io.NopCloser(bytes.NewReader(responseBodyData)),
		ContentLength: int64(len(responseBodyData)),
	}

	return resp, nil
}

// Transport implements http.RoundTripper interface using go-curl-impersonate.
// It provides browser impersonation capabilities while maintaining full
// compatibility with the standard http.RoundTripper interface.
type Transport struct {
	// ImpersonateTarget specifies which browser to impersonate (e.g., "chrome136").
	// Supported targets: chrome136, firefox102, safari17_0, edge122
	ImpersonateTarget string

	// UseDefaultHeaders whether to use default headers for the impersonated browser.
	UseDefaultHeaders bool
}

// NewTransport creates a new Transport with default settings
func NewTransport() *Transport {
	return &Transport{
		ImpersonateTarget: "chrome136",
		UseDefaultHeaders: true,
	}
}

// RoundTrip executes a single HTTP transaction using go-curl-impersonate.
// It implements the http.RoundTripper interface and provides browser impersonation.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if req.URL == nil {
		return nil, fmt.Errorf("request URL cannot be nil")
	}

	// Convert headers to simple map
	headers := make(map[string]string)
	for name, values := range req.Header {
		if len(values) > 0 {
			headers[name] = values[0] // Take first value for simplicity
		}
	}

	// Read request body if present
	var body []byte
	if req.Body != nil {
		var err error
		body, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		req.Body.Close()
	}

	// Perform the request using our working helper function
	resp, err := performRequest(req.URL.String(), req.Method, headers, body)
	if err != nil {
		return nil, err
	}

	// Set the request reference
	resp.Request = req

	return resp, nil
}

// parseHeaders parses raw HTTP headers into http.Header
func parseHeaders(headerData string) http.Header {
	headers := make(http.Header)
	lines := strings.Split(headerData, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip the status line (HTTP/1.1 200 OK)
		if strings.HasPrefix(line, "HTTP/") {
			continue
		}

		// Parse header line (name: value)
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			headers.Add(name, value)
		}
	}

	return headers
}

// Client wraps http.Client to use our custom Transport that provides
// browser impersonation capabilities. It embeds http.Client so all
// standard methods are available.
type Client struct {
	*http.Client
}

// NewClient creates a new Client that uses go-curl-impersonate with default settings.
// The client will impersonate Chrome 136 by default with a 30-second timeout.
func NewClient() *Client {
	return &Client{
		Client: &http.Client{
			Transport: NewTransport(),
			Timeout:   30 * time.Second,
		},
	}
}

// NewClientWithTarget creates a new Client with a specific impersonation target.
// Supported targets include: chrome136, firefox102, safari17_0, edge122.
func NewClientWithTarget(target string) *Client {
	return &Client{
		Client: &http.Client{
			Transport: &Transport{
				ImpersonateTarget: target,
				UseDefaultHeaders: true,
			},
			Timeout: 30 * time.Second,
		},
	}
}

// DefaultClient is the default client that uses curl-impersonate
// This allows drop-in compatibility with net/http package-level functions
var DefaultClient = &Client{
	Client: &http.Client{
		Transport: NewTransport(),
		Timeout:   30 * time.Second,
	},
}

// Package-level functions for drop-in compatibility with net/http

// Get makes a GET request using the default client
func Get(url string) (*Response, error) {
	return DefaultClient.Get(url)
}

// Post makes a POST request using the default client
func Post(url, contentType string, body io.Reader) (*Response, error) {
	return DefaultClient.Post(url, contentType, body)
}

// PostForm posts a form using the default client
func PostForm(url string, data url.Values) (*Response, error) {
	return DefaultClient.PostForm(url, data)
}

// Head makes a HEAD request using the default client
func Head(url string) (*Response, error) {
	return DefaultClient.Head(url)
}

// Do executes a request using the default client
func Do(req *Request) (*Response, error) {
	return DefaultClient.Do(req)
}
