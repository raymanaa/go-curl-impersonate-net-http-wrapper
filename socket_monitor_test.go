package curlhttp

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"
)

// SocketStats represents socket connection statistics
type SocketStats struct {
	Established int
	TimeWait    int
	CloseWait   int
	Listen      int
	Total       int
}

// getSocketStats parses netstat output to count socket states
func getSocketStats() (SocketStats, error) {
	cmd := exec.Command("netstat", "-an")
	output, err := cmd.Output()
	if err != nil {
		return SocketStats{}, fmt.Errorf("failed to run netstat: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	stats := SocketStats{}

	// Regex patterns for socket states
	establishedRe := regexp.MustCompile(`\s+ESTABLISHED\s*$`)
	timeWaitRe := regexp.MustCompile(`\s+TIME_WAIT\s*$`)
	closeWaitRe := regexp.MustCompile(`\s+CLOSE_WAIT\s*$`)
	listenRe := regexp.MustCompile(`\s+LISTEN\s*$`)

	for _, line := range lines {
		if strings.Contains(line, "tcp") {
			stats.Total++

			if establishedRe.MatchString(line) {
				stats.Established++
			} else if timeWaitRe.MatchString(line) {
				stats.TimeWait++
			} else if closeWaitRe.MatchString(line) {
				stats.CloseWait++
			} else if listenRe.MatchString(line) {
				stats.Listen++
			}
		}
	}

	return stats, nil
}

// TestSocketStateMonitoring monitors socket states during HTTP requests
func TestSocketStateMonitoring(t *testing.T) {
	// Get baseline socket stats
	baseline, err := getSocketStats()
	if err != nil {
		t.Logf("Warning: Could not get baseline socket stats: %v", err)
		baseline = SocketStats{}
	}

	t.Logf("Baseline sockets - Total: %d, Established: %d, TIME_WAIT: %d, CLOSE_WAIT: %d",
		baseline.Total, baseline.Established, baseline.TimeWait, baseline.CloseWait)

	// Create mock server
	server := createMockServer()
	defer server.Close()

	// Create client
	client := NewClient()
	client.Timeout = 2 * time.Second

	const numRequests = 100

	t.Logf("Making %d HTTP requests and monitoring socket states...", numRequests)

	for i := 0; i < numRequests; i++ {
		// Make a request
		resp, err := client.Get(fmt.Sprintf("%s/get?id=%d", server.URL, i))
		if err != nil {
			t.Logf("Request %d failed: %v", i, err)
			continue
		}

		// Read and close response
		_, readErr := resp.Body.Read(make([]byte, 1024))
		resp.Body.Close()

		if readErr != nil && readErr.Error() != "EOF" {
			t.Logf("Read error on request %d: %v", i, readErr)
		}

		// Monitor socket states every 10 requests
		if (i+1)%10 == 0 {
			stats, statErr := getSocketStats()
			if statErr == nil {
				deltaTotal := stats.Total - baseline.Total
				deltaEstablished := stats.Established - baseline.Established
				deltaTimeWait := stats.TimeWait - baseline.TimeWait
				deltaCloseWait := stats.CloseWait - baseline.CloseWait

				t.Logf("After %d requests - Delta sockets: Total: +%d, Established: +%d, TIME_WAIT: +%d, CLOSE_WAIT: +%d",
					i+1, deltaTotal, deltaEstablished, deltaTimeWait, deltaCloseWait)

				// Warning if TIME_WAIT is growing rapidly
				if deltaTimeWait > 50 {
					t.Logf("⚠️  WARNING: TIME_WAIT sockets growing rapidly: +%d", deltaTimeWait)
				}
			}
		}

		// Small delay to allow socket state changes to propagate
		time.Sleep(10 * time.Millisecond)
	}

	// Final socket stats
	final, err := getSocketStats()
	if err == nil {
		deltaTotal := final.Total - baseline.Total
		deltaTimeWait := final.TimeWait - baseline.TimeWait
		deltaEstablished := final.Established - baseline.Established

		t.Logf("=== FINAL SOCKET ANALYSIS ===")
		t.Logf("Total socket delta: +%d", deltaTotal)
		t.Logf("TIME_WAIT delta: +%d", deltaTimeWait)
		t.Logf("ESTABLISHED delta: +%d", deltaEstablished)

		// Test for socket leakage - warn if excessive but don't fail
		if deltaTimeWait > numRequests/2 {
			t.Logf("⚠️  High TIME_WAIT socket count: %d (requests: %d) - connections not being reused efficiently", deltaTimeWait, numRequests)
		}

		if deltaEstablished > 5 {
			t.Errorf("Socket leakage detected: %d ESTABLISHED sockets still open", deltaEstablished)
		}
	}
}
