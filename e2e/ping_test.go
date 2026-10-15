//go:build e2e && (linux || windows)

package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

type pingResponse struct {
	SourceIP  *string    `json:"source_ip"`
	LatencyMS *int64     `json:"latency_ms"`
	Time      *time.Time `json:"time"`
}

// TestPing Builds the real binary, starts it and calls /ping over HTTP
func TestPing(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "latency-service")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}

	build := exec.Command("go", "build", "-o", binary, "../cmd")
	build.Stdout, build.Stderr = os.Stdout, os.Stderr
	if err := build.Run(); err != nil {
		t.Fatalf("build failed: %s", err)
	}

	addr := freeAddr(t)
	server := exec.Command(binary)
	server.Env = append(os.Environ(),
		"LATENCY_LISTEN_HTTPS="+addr,
		"LATENCY_LOGGING=false",
		"LATENCY_DATA_DIRECTORY="+t.TempDir(),
	)
	server.Stdout, server.Stderr = os.Stdout, os.Stderr
	if err := server.Start(); err != nil {
		t.Fatalf("starting server failed: %s", err)
	}
	t.Cleanup(func() {
		_ = server.Process.Kill()
		_ = server.Wait()
	})

	url := fmt.Sprintf("http://%s/ping", addr)
	resp := waitForServer(t, url)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", resp.StatusCode, body)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}

	var ping pingResponse
	if err := json.Unmarshal(body, &ping); err != nil {
		t.Fatalf("invalid json %q: %s", body, err)
	}
	if ping.SourceIP == nil || *ping.SourceIP != "127.0.0.1" {
		t.Fatalf("expected source_ip 127.0.0.1, got %s", body)
	}
	if ping.LatencyMS == nil || *ping.LatencyMS < 0 {
		t.Fatalf("expected non-negative latency_ms, got %s", body)
	}
	if ping.Time == nil || ping.Time.IsZero() {
		t.Fatalf("expected time, got %s", body)
	}
	t.Logf("response: %s", body)
}

func freeAddr(t *testing.T) string {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().String()
}

func waitForServer(t *testing.T, url string) *http.Response {
	deadline := time.Now().Add(30 * time.Second)
	for {
		resp, err := http.Get(url)
		if err == nil {
			return resp
		}
		if time.Now().After(deadline) {
			t.Fatalf("server did not come up: %s", err)
		}
		time.Sleep(200 * time.Millisecond)
	}
}
