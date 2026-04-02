package auth

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/larksuite/cli/internal/core"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

var (
	authResponseLogWriter io.Writer = defaultLogWriter{}
	authResponseLogNow              = time.Now
	authResponseLogArgs             = func() []string { return os.Args }

	logMu sync.Mutex
)

type defaultLogWriter struct{}

func (defaultLogWriter) Write(p []byte) (n int, err error) {
	logMu.Lock()
	defer logMu.Unlock()

	dir := filepath.Join(core.GetConfigDir(), "logs")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return 0, err
	}

	// Format: auth-2006-01-02.log
	now := authResponseLogNow()
	logName := fmt.Sprintf("auth-%s.log", now.Format("2006-01-02"))
	logPath := filepath.Join(dir, logName)

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	// Clean up old logs (keep 7 days)
	go cleanupOldLogs(dir, now)

	return f.Write(p)
}

func cleanupOldLogs(dir string, now time.Time) {
	defer func() {
		if r := recover(); r != nil {
			// Record the panic so we can debug without crashing the main program
			msg := fmt.Sprintf("[lark-cli] [WARN] background log cleanup panicked: %v\n", r)
			_, _ = authResponseLogWriter.Write([]byte(msg))
		}
	}()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	cutoff := now.AddDate(0, 0, -7)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "auth-") || !strings.HasSuffix(entry.Name(), ".log") {
			continue
		}

		// Extract date from filename
		dateStr := strings.TrimPrefix(entry.Name(), "auth-")
		dateStr = strings.TrimSuffix(dateStr, ".log")

		logDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		// If log is older than 7 days, delete it
		if logDate.Before(cutoff) {
			_ = os.Remove(filepath.Join(dir, entry.Name()))
		}
	}
}

type authResponse interface {
	logStatusCode() int
	logHeader(key string) string
	logPath() string
}

type httpAuthResponse struct {
	*http.Response
}

func (r httpAuthResponse) logStatusCode() int {
	return r.StatusCode
}

func (r httpAuthResponse) logHeader(key string) string {
	return r.Header.Get(key)
}

func (r httpAuthResponse) logPath() string {
	if r.Request != nil && r.Request.URL != nil {
		return r.Request.URL.Path
	}
	return "miss"
}

type sdkAuthResponse struct {
	path string
	*larkcore.ApiResp
}

func (r sdkAuthResponse) logStatusCode() int {
	return r.StatusCode
}

func (r sdkAuthResponse) logHeader(key string) string {
	return r.Header.Get(key)
}

func (r sdkAuthResponse) logPath() string {
	return r.path
}

func logAuthResponse(resp interface{}) {
	if authResponseLogWriter == nil || resp == nil {
		return
	}

	var r authResponse
	switch v := resp.(type) {
	case *http.Response:
		r = httpAuthResponse{v}
	case authResponse:
		r = v
	default:
		return
	}

	fmt.Fprintf(
		authResponseLogWriter,
		"[lark-cli] auth-response: time=%s path=%s status=%d x-tt-logid=%s cmdline=%s\n",
		authResponseLogNow().Format(time.RFC3339Nano),
		r.logPath(),
		r.logStatusCode(),
		r.logHeader("x-tt-logid"),
		strings.Join(authResponseLogArgs(), " "),
	)
}
