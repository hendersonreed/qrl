//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const serverAddr = "http://localhost:8080"

var serverCmd *exec.Cmd

func TestMain(m *testing.M) {
	serverCmd = exec.Command("./qrl")
	serverCmd.Dir = "../"
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr

	if err := serverCmd.Start(); err != nil {
		panic("failed to start server: " + err.Error())
	}

	if err := waitForServer(); err != nil {
		_ = serverCmd.Process.Kill()
		panic("server failed to start: " + err.Error())
	}

	code := m.Run()

	_ = serverCmd.Process.Kill()
	_ = serverCmd.Wait()

	if code != 0 {
		panic(code)
	}
}

func waitForServer() error {
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(serverAddr + "/")
		if err == nil {
			resp.Body.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return context.DeadlineExceeded
}

func TestGETIndex(t *testing.T) {
	resp, err := http.Get(serverAddr + "/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("expected text/html; charset=utf-8, got %s", ct)
	}
}

func TestPUTQRCode(t *testing.T) {
	payload := []byte("hello world")
	resp, err := http.DefaultClient.Do(&http.Request{
		Method: http.MethodPut,
		URL:    mustParse(serverAddr + "/"),
		Body:   io.NopCloser(bytes.NewReader(payload)),
	})
	if err != nil {
		t.Fatalf("PUT / failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "image/png" {
		t.Errorf("expected image/png, got %s", ct)
	}

	data, _ := io.ReadAll(resp.Body)
	if len(data) == 0 {
		t.Error("expected non-empty QR code image")
	}
}

func TestUnsupportedMethod(t *testing.T) {
	resp, err := http.Post(serverAddr+"/", "text/plain", bytes.NewBufferString("data"))
	if err != nil {
		t.Fatalf("POST / failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 Method Not Allowed, got %d", resp.StatusCode)
	}
}

// --- New tests ---

func TestPUTEmptyInput(t *testing.T) {
	resp := putRequest(t, "")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for empty input, got %d", resp.StatusCode)
	}
}

func TestPUTWhitespaceInput(t *testing.T) {
	resp := putRequest(t, " \n\t\r ")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for whitespace input, got %d", resp.StatusCode)
	}
}

func TestPUTInvalidUTF8Input(t *testing.T) {
	payload := []byte{0xff, 0xfe, 0xfd}
	resp := putRawRequest(t, payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for invalid UTF-8 input, got %d", resp.StatusCode)
	}
}

func TestPUTControlCharsInput(t *testing.T) {
	payload := "Hello\x01World"
	resp := putRequest(t, payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for control chars input, got %d", resp.StatusCode)
	}
}

func TestPUTAllowedControlCharsInput(t *testing.T) {
	payload := "Line1\nLine2\r\nLine3\tEnd"
	resp := putRequest(t, payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK for allowed control chars input, got %d", resp.StatusCode)
	}
}

func TestPUTExceedsMaxLength(t *testing.T) {
	payload := strings.Repeat("a", 3001) // 1 byte over limit
	resp := putRequest(t, payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413 Request Entity Too Large for oversized input, got %d", resp.StatusCode)
	}
}

// --- helpers ---

func putRequest(t *testing.T, body string) *http.Response {
	t.Helper()
	return putRawRequest(t, []byte(body))
}

func putRawRequest(t *testing.T, body []byte) *http.Response {
	t.Helper()
	resp, err := http.DefaultClient.Do(&http.Request{
		Method: http.MethodPut,
		URL:    mustParse(serverAddr + "/"),
		Body:   io.NopCloser(bytes.NewReader(body)),
	})
	if err != nil {
		t.Fatalf("PUT / failed: %v", err)
	}
	return resp
}

func mustParse(raw string) *url.URL {
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		panic(err)
	}
	return u
}
