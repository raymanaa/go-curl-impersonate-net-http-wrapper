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
//	import http "github.com/BridgeSenseDev/go-curl-impersonate-net-http-wrapper"
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
	CloseNotifier  = http.CloseNotifier
	Pusher         = http.Pusher
	ConnState      = http.ConnState
	Dir            = http.Dir
	ProtocolError  = http.ProtocolError
	SameSite       = http.SameSite
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

// Re-export all status codes for complete compatibility
const (
	// 1xx Informational
	StatusContinue           = http.StatusContinue           // 100
	StatusSwitchingProtocols = http.StatusSwitchingProtocols // 101
	StatusProcessing         = http.StatusProcessing         // 102
	StatusEarlyHints         = http.StatusEarlyHints         // 103

	// 2xx Success
	StatusOK                   = http.StatusOK                   // 200
	StatusCreated              = http.StatusCreated              // 201
	StatusAccepted             = http.StatusAccepted             // 202
	StatusNonAuthoritativeInfo = http.StatusNonAuthoritativeInfo // 203
	StatusNoContent            = http.StatusNoContent            // 204
	StatusResetContent         = http.StatusResetContent         // 205
	StatusPartialContent       = http.StatusPartialContent       // 206
	StatusMultiStatus          = http.StatusMultiStatus          // 207
	StatusAlreadyReported      = http.StatusAlreadyReported      // 208
	StatusIMUsed               = http.StatusIMUsed               // 226

	// 3xx Redirection
	StatusMultipleChoices   = http.StatusMultipleChoices   // 300
	StatusMovedPermanently  = http.StatusMovedPermanently  // 301
	StatusFound             = http.StatusFound             // 302
	StatusSeeOther          = http.StatusSeeOther          // 303
	StatusNotModified       = http.StatusNotModified       // 304
	StatusUseProxy          = http.StatusUseProxy          // 305
	StatusTemporaryRedirect = http.StatusTemporaryRedirect // 307
	StatusPermanentRedirect = http.StatusPermanentRedirect // 308

	// 4xx Client Errors
	StatusBadRequest                   = http.StatusBadRequest                   // 400
	StatusUnauthorized                 = http.StatusUnauthorized                 // 401
	StatusPaymentRequired              = http.StatusPaymentRequired              // 402
	StatusForbidden                    = http.StatusForbidden                    // 403
	StatusNotFound                     = http.StatusNotFound                     // 404
	StatusMethodNotAllowed             = http.StatusMethodNotAllowed             // 405
	StatusNotAcceptable                = http.StatusNotAcceptable                // 406
	StatusProxyAuthRequired            = http.StatusProxyAuthRequired            // 407
	StatusRequestTimeout               = http.StatusRequestTimeout               // 408
	StatusConflict                     = http.StatusConflict                     // 409
	StatusGone                         = http.StatusGone                         // 410
	StatusLengthRequired               = http.StatusLengthRequired               // 411
	StatusPreconditionFailed           = http.StatusPreconditionFailed           // 412
	StatusRequestEntityTooLarge        = http.StatusRequestEntityTooLarge        // 413
	StatusRequestURITooLong            = http.StatusRequestURITooLong            // 414
	StatusUnsupportedMediaType         = http.StatusUnsupportedMediaType         // 415
	StatusRequestedRangeNotSatisfiable = http.StatusRequestedRangeNotSatisfiable // 416
	StatusExpectationFailed            = http.StatusExpectationFailed            // 417
	StatusTeapot                       = http.StatusTeapot                       // 418
	StatusMisdirectedRequest           = http.StatusMisdirectedRequest           // 421
	StatusUnprocessableEntity          = http.StatusUnprocessableEntity          // 422
	StatusLocked                       = http.StatusLocked                       // 423
	StatusFailedDependency             = http.StatusFailedDependency             // 424
	StatusTooEarly                     = http.StatusTooEarly                     // 425
	StatusUpgradeRequired              = http.StatusUpgradeRequired              // 426
	StatusPreconditionRequired         = http.StatusPreconditionRequired         // 428
	StatusTooManyRequests              = http.StatusTooManyRequests              // 429
	StatusRequestHeaderFieldsTooLarge  = http.StatusRequestHeaderFieldsTooLarge  // 431
	StatusUnavailableForLegalReasons   = http.StatusUnavailableForLegalReasons   // 451

	// 5xx Server Errors
	StatusInternalServerError           = http.StatusInternalServerError           // 500
	StatusNotImplemented                = http.StatusNotImplemented                // 501
	StatusBadGateway                    = http.StatusBadGateway                    // 502
	StatusServiceUnavailable            = http.StatusServiceUnavailable            // 503
	StatusGatewayTimeout                = http.StatusGatewayTimeout                // 504
	StatusHTTPVersionNotSupported       = http.StatusHTTPVersionNotSupported       // 505
	StatusVariantAlsoNegotiates         = http.StatusVariantAlsoNegotiates         // 506
	StatusInsufficientStorage           = http.StatusInsufficientStorage           // 507
	StatusLoopDetected                  = http.StatusLoopDetected                  // 508
	StatusNotExtended                   = http.StatusNotExtended                   // 510
	StatusNetworkAuthenticationRequired = http.StatusNetworkAuthenticationRequired // 511
)

// Re-export common constants
const (
	DefaultMaxHeaderBytes      = http.DefaultMaxHeaderBytes      // 1048576
	DefaultMaxIdleConnsPerHost = http.DefaultMaxIdleConnsPerHost // 2
	TimeFormat                 = http.TimeFormat                 // "Mon, 02 Jan 2006 15:04:05 GMT"
	TrailerPrefix              = http.TrailerPrefix              // "Trailer:"
)

// Re-export error variables for complete compatibility
var (
	ErrBodyNotAllowed     = http.ErrBodyNotAllowed
	ErrBodyReadAfterClose = http.ErrBodyReadAfterClose
	ErrHandlerTimeout     = http.ErrHandlerTimeout
	ErrLineTooLong        = http.ErrLineTooLong
	ErrMissingFile        = http.ErrMissingFile
	ErrNoCookie           = http.ErrNoCookie
	ErrNoLocation         = http.ErrNoLocation
	ErrSchemeMismatch     = http.ErrSchemeMismatch
	ErrServerClosed       = http.ErrServerClosed
	ErrSkipAltProtocol    = http.ErrSkipAltProtocol
	ErrUseLastResponse    = http.ErrUseLastResponse
	ErrAbortHandler       = http.ErrAbortHandler
	NoBody                = http.NoBody
)

// Re-export context keys
var (
	ServerContextKey = http.ServerContextKey
)

// Re-export common functions
var (
	NewRequest            = http.NewRequest
	NewRequestWithContext = http.NewRequestWithContext
	ReadResponse          = http.ReadResponse
	ParseHTTPVersion      = http.ParseHTTPVersion
	ParseTime             = http.ParseTime
	ParseCookie           = http.ParseCookie
	ParseSetCookie        = http.ParseSetCookie
	StatusText            = http.StatusText
	CanonicalHeaderKey    = http.CanonicalHeaderKey
	DetectContentType     = http.DetectContentType
	MaxBytesReader        = http.MaxBytesReader
	Error                 = http.Error
	NotFound              = http.NotFound
	Redirect              = http.Redirect
	Serve                 = http.Serve
	ServeContent          = http.ServeContent
	ServeFile             = http.ServeFile
	SetCookie             = http.SetCookie
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

// responseBuffer is a thread-safe buffer for collecting response data in memory
type responseBuffer struct {
	buffer *bytes.Buffer
	mu     sync.Mutex
}

// Write implements io.Writer for thread-safe writing
func (rb *responseBuffer) Write(p []byte) (int, error) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.buffer.Write(p)
}

// Bytes returns a copy of the buffer contents
func (rb *responseBuffer) Bytes() []byte {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.buffer.Bytes()
}

// Reset clears the buffer
func (rb *responseBuffer) Reset() {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.buffer.Reset()
}

// writeDataToBuffer is the callback function for writing response data to a buffer
func writeDataToBuffer(ptr []byte, userdata interface{}) bool {
	buffer, ok := userdata.(*responseBuffer)
	if !ok {
		return false
	}
	_, err := buffer.Write(ptr)
	return err == nil
}

// Transport implements http.RoundTripper interface using go-curl-impersonate.
// It provides browser impersonation capabilities while maintaining full
// compatibility with the standard http.RoundTripper interface.
// Now includes connection pooling and in-memory responses for optimal performance.
type Transport struct {
	// ImpersonateTarget specifies which browser to impersonate (e.g., "chrome136").
	// Supported targets: chrome136, firefox102, safari17_0, edge122
	ImpersonateTarget string

	// UseDefaultHeaders whether to use default headers for the impersonated browser.
	UseDefaultHeaders bool

	// Connection pooling for performance
	curlHandles chan *curl.CURL
	maxPoolSize int
	poolOnce    sync.Once
}

// initPool initializes the connection pool for the transport
func (t *Transport) initPool() {
	t.poolOnce.Do(func() {
		if t.maxPoolSize == 0 {
			t.maxPoolSize = 10 // Default pool size
		}
		t.curlHandles = make(chan *curl.CURL, t.maxPoolSize)
	})
}

// getCurlHandle gets a curl handle from the pool or creates a new one
func (t *Transport) getCurlHandle() *curl.CURL {
	t.initPool()

	select {
	case handle := <-t.curlHandles:
		return handle
	default:
		// No available handle, create new one
		initCurl()
		easy := curl.EasyInit()
		if easy == nil {
			return nil
		}

		// Set up persistent options for performance
		easy.Setopt(curl.OPT_HEADER, true)
		easy.Setopt(curl.OPT_NOPROGRESS, true)
		if t.ImpersonateTarget == "" {
			t.ImpersonateTarget = "chrome136" // Default target
		}
		easy.Impersonate(t.ImpersonateTarget, t.UseDefaultHeaders)

		// Enable connection reuse and keep-alive for better performance
		easy.Setopt(curl.OPT_FRESH_CONNECT, false)
		easy.Setopt(curl.OPT_FORBID_REUSE, false)
		easy.Setopt(curl.OPT_TCP_KEEPALIVE, true)
		easy.Setopt(curl.OPT_TCP_KEEPIDLE, 60)
		easy.Setopt(curl.OPT_TCP_KEEPINTVL, 60)

		return easy
	}
}

// returnCurlHandle returns a handle to the pool for reuse
func (t *Transport) returnCurlHandle(handle *curl.CURL) {
	if handle == nil {
		return
	}

	// Reset handle for reuse (but keep connection alive)
	handle.Reset()

	// Set up persistent options again after reset
	handle.Setopt(curl.OPT_HEADER, true)
	handle.Setopt(curl.OPT_NOPROGRESS, true)
	handle.Impersonate(t.ImpersonateTarget, t.UseDefaultHeaders)
	handle.Setopt(curl.OPT_FRESH_CONNECT, false)
	handle.Setopt(curl.OPT_FORBID_REUSE, false)
	handle.Setopt(curl.OPT_TCP_KEEPALIVE, true)
	handle.Setopt(curl.OPT_TCP_KEEPIDLE, 60)
	handle.Setopt(curl.OPT_TCP_KEEPINTVL, 60)

	select {
	case t.curlHandles <- handle:
		// Successfully returned to pool
	default:
		// Pool is full, cleanup the handle
		handle.Cleanup()
	}
}

// NewTransport creates a new Transport with default settings and connection pooling
func NewTransport() *Transport {
	return &Transport{
		ImpersonateTarget: "chrome136",
		UseDefaultHeaders: true,
		maxPoolSize:       10, // Default pool size
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

	// Use optimized request with connection pooling and in-memory responses
	resp, err := t.performOptimizedRequest(req.URL.String(), req.Method, headers, body)
	if err != nil {
		return nil, err
	}

	// Set the request reference
	resp.Request = req

	return resp, nil
}

// performOptimizedRequest performs HTTP request using in-memory buffer and connection pooling
func (t *Transport) performOptimizedRequest(url, method string, headers map[string]string, body []byte) (*http.Response, error) {
	// Get curl handle from pool
	easy := t.getCurlHandle()
	if easy == nil {
		return nil, fmt.Errorf("failed to get curl handle")
	}
	defer t.returnCurlHandle(easy)

	// Set the URL
	if err := easy.Setopt(curl.OPT_URL, url); err != nil {
		return nil, fmt.Errorf("failed to set URL: %w", err)
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
		if len(body) > 0 {
			if err := easy.Setopt(curl.OPT_POSTFIELDS, body); err != nil {
				return nil, fmt.Errorf("failed to set request body: %w", err)
			}
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

	// Create in-memory response buffer instead of temporary file
	responseBuffer := &responseBuffer{
		buffer: bytes.NewBuffer(make([]byte, 0, 4096)), // Pre-allocate 4KB
	}

	// Set response callback function with buffer as userdata
	if err := easy.Setopt(curl.OPT_WRITEFUNCTION, writeDataToBuffer); err != nil {
		return nil, fmt.Errorf("failed to set write function: %w", err)
	}
	if err := easy.Setopt(curl.OPT_WRITEDATA, responseBuffer); err != nil {
		return nil, fmt.Errorf("failed to set write data: %w", err)
	}

	// Perform the request
	if err := easy.Perform(); err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Get response code
	responseCodeInfo, err := easy.Getinfo(uint32(curl.CURLINFO_RESPONSE_CODE))
	if err != nil {
		return nil, fmt.Errorf("failed to get response code: %w", err)
	}
	responseCode := int(responseCodeInfo.(int64))

	// Get the combined response (headers + body) from buffer
	combinedData := responseBuffer.Bytes()

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
	http.Client
	initialized bool
}

// ensureInitialized initializes the Client with default settings if it hasn't been initialized yet.
// This enables zero-value Client instances to work properly for drop-in compatibility.
func (c *Client) ensureInitialized() {
	if !c.initialized {
		if c.Transport == nil {
			c.Transport = NewTransport()
		}
		if c.Timeout == 0 {
			c.Timeout = 30 * time.Second
		}
		c.initialized = true
	}
}

// NewClient creates a new Client that uses go-curl-impersonate with default settings.
// The client will impersonate Chrome 136 by default with a 30-second timeout.
func NewClient() *Client {
	return &Client{
		Client: http.Client{
			Transport: NewTransport(),
			Timeout:   30 * time.Second,
		},
		initialized: true,
	}
}

// NewClientWithTarget creates a new Client with a specific impersonation target.
// Supported targets include: chrome136, firefox102, safari17_0, edge122.
func NewClientWithTarget(target string) *Client {
	return &Client{
		Client: http.Client{
			Transport: &Transport{
				ImpersonateTarget: target,
				UseDefaultHeaders: true,
			},
			Timeout: 30 * time.Second,
		},
		initialized: true,
	}
}

// DefaultClient is the default client that uses curl-impersonate
// This allows drop-in compatibility with net/http package-level functions
var DefaultClient = &Client{
	Client: http.Client{
		Transport: NewTransport(),
		Timeout:   30 * time.Second,
	},
	initialized: true,
}

// Override key methods to ensure initialization for zero-value clients

// Get makes a GET request. Ensures the client is initialized if needed for zero-value compatibility.
func (c *Client) Get(url string) (*Response, error) {
	c.ensureInitialized()
	return c.Client.Get(url)
}

// Post makes a POST request. Ensures the client is initialized if needed for zero-value compatibility.
func (c *Client) Post(url, contentType string, body io.Reader) (*Response, error) {
	c.ensureInitialized()
	return c.Client.Post(url, contentType, body)
}

// PostForm makes a POST request with form data. Ensures the client is initialized if needed for zero-value compatibility.
func (c *Client) PostForm(url string, data url.Values) (*Response, error) {
	c.ensureInitialized()
	return c.Client.PostForm(url, data)
}

// Head makes a HEAD request. Ensures the client is initialized if needed for zero-value compatibility.
func (c *Client) Head(url string) (*Response, error) {
	c.ensureInitialized()
	return c.Client.Head(url)
}

// Do sends an HTTP request. Ensures the client is initialized if needed for zero-value compatibility.
func (c *Client) Do(req *Request) (*Response, error) {
	c.ensureInitialized()
	return c.Client.Do(req)
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
