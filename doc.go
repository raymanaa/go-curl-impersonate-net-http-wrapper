/*
Package curlhttp provides a drop-in replacement for net/http that uses
curl-impersonate for browser impersonation to avoid detection by websites
that block automated requests.

This package re-exports all net/http types, constants, and functions while
using curl-impersonate as the underlying HTTP client. Simply replace your
net/http import with this package for seamless browser impersonation.

# Basic Usage

The simplest way to use this package is to replace your net/http import:

	// Instead of: import "net/http"
	import http "github.com/BridgeSenseDev/go-curl-impersonate-net-http-wrapper"

	// All your existing net/http code works unchanged!
	resp, err := http.Get("https://example.com")

# Using the Client

You can also create clients directly for more control:

	client := curlhttp.NewClient()
	resp, err := client.Get("https://example.com")

	// Or with a specific browser target:
	client := curlhttp.NewClientWithTarget("firefox102")
	resp, err := client.Get("https://example.com")

# Browser Impersonation

By default, the package impersonates Chrome 136. Supported targets include:
- chrome136 (default)
- firefox102
- safari17_0
- edge122

The impersonation includes proper TLS fingerprints and headers to avoid detection.

# Compatibility

This package provides 100% API compatibility with net/http. All types, constants,
and functions are re-exported so existing code works without modification.

# Performance

The wrapper adds minimal overhead over net/http while providing powerful browser
impersonation capabilities. Response times are typically within 10-50ms of
standard net/http requests.
*/
package curlhttp
