// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package auth

import (
	"bytes"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/larksuite/cli/internal/core"
	"github.com/larksuite/cli/internal/httpmock"
)

func TestResolveOAuthEndpoints_Feishu(t *testing.T) {
	ep := ResolveOAuthEndpoints(core.BrandFeishu)
	if ep.DeviceAuthorization != "https://accounts.feishu.cn/oauth/v1/device_authorization" {
		t.Errorf("DeviceAuthorization = %q", ep.DeviceAuthorization)
	}
	if ep.Token != "https://open.feishu.cn/open-apis/authen/v2/oauth/token" {
		t.Errorf("Token = %q", ep.Token)
	}
}

func TestResolveOAuthEndpoints_Lark(t *testing.T) {
	ep := ResolveOAuthEndpoints(core.BrandLark)
	if ep.DeviceAuthorization != "https://accounts.larksuite.com/oauth/v1/device_authorization" {
		t.Errorf("DeviceAuthorization = %q", ep.DeviceAuthorization)
	}
	if ep.Token != "https://open.larksuite.com/open-apis/authen/v2/oauth/token" {
		t.Errorf("Token = %q", ep.Token)
	}
}

func TestRequestDeviceAuthorization_LogsResponse(t *testing.T) {
	reg := &httpmock.Registry{}
	t.Cleanup(func() { reg.Verify(t) })

	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    PathDeviceAuthorization,
		Body: map[string]interface{}{
			"device_code":               "device-code",
			"user_code":                 "user-code",
			"verification_uri":          "https://example.com/verify",
			"verification_uri_complete": "https://example.com/verify?code=123",
			"expires_in":                240,
			"interval":                  5,
		},
		Headers: http.Header{
			"Content-Type": []string{"application/json"},
			"X-Tt-Logid":   []string{"device-log-id"},
		},
	})

	var buf bytes.Buffer
	prevWriter := authResponseLogWriter
	prevNow := authResponseLogNow
	prevArgs := authResponseLogArgs
	authResponseLogWriter = &buf
	authResponseLogNow = func() time.Time {
		return time.Date(2026, 4, 2, 3, 4, 5, 0, time.UTC)
	}
	authResponseLogArgs = func() []string {
		return []string{"lark-cli", "auth", "login"}
	}
	t.Cleanup(func() {
		authResponseLogWriter = prevWriter
		authResponseLogNow = prevNow
		authResponseLogArgs = prevArgs
	})

	_, err := RequestDeviceAuthorization(httpmock.NewClient(reg), "cli_a", "secret_b", core.BrandFeishu, "", nil)
	if err != nil {
		t.Fatalf("RequestDeviceAuthorization() error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "time=2026-04-02T03:04:05Z") {
		t.Fatalf("expected time in log, got %q", got)
	}
	if !strings.Contains(got, "path=miss") {
		t.Fatalf("expected path in log, got %q", got)
	}
	if !strings.Contains(got, "status=200") {
		t.Fatalf("expected status=200 in log, got %q", got)
	}
	if !strings.Contains(got, "x-tt-logid=device-log-id") {
		t.Fatalf("expected x-tt-logid in log, got %q", got)
	}
	if !strings.Contains(got, "cmdline=lark-cli auth login") {
		t.Fatalf("expected cmdline in log, got %q", got)
	}
}
