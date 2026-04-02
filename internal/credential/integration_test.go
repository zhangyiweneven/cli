package credential_test

import (
	"context"
	"testing"

	extcred "github.com/larksuite/cli/extension/credential"
	envprovider "github.com/larksuite/cli/extension/credential/env"
	"github.com/larksuite/cli/internal/core"
	"github.com/larksuite/cli/internal/credential"
)

func TestFullChain_EnvWins(t *testing.T) {
	t.Setenv("LARK_APP_ID", "env_app")
	t.Setenv("LARK_APP_SECRET", "env_secret")
	t.Setenv("LARK_USER_ACCESS_TOKEN", "env_uat")

	ep := &envprovider.Provider{}
	cp := credential.NewCredentialProvider(
		[]extcred.Provider{ep},
		nil, nil, nil,
	)

	acct, err := cp.ResolveAccount(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if acct.AppID != "env_app" {
		t.Errorf("expected env_app, got %s", acct.AppID)
	}

	result, err := cp.ResolveToken(context.Background(), credential.TokenSpec{
		Type: credential.TokenTypeUAT, Identity: core.AsUser, AppID: "env_app",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Token != "env_uat" {
		t.Errorf("expected env_uat, got %s", result.Token)
	}
}

func TestFullChain_Fallthrough(t *testing.T) {
	// env provider returns nil (no env vars set), falls through to default token
	ep := &envprovider.Provider{}
	mock := &mockDefaultTokenProvider{token: "mock_tok", scopes: "drive:read"}

	cp := credential.NewCredentialProvider(
		[]extcred.Provider{ep},
		nil, mock, nil,
	)
	result, err := cp.ResolveToken(context.Background(), credential.TokenSpec{
		Type: credential.TokenTypeUAT, Identity: core.AsUser, AppID: "app1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Token != "mock_tok" || result.Scopes != "drive:read" {
		t.Errorf("unexpected: %+v", result)
	}
}

type mockDefaultTokenProvider struct {
	token  string
	scopes string
}

func (m *mockDefaultTokenProvider) ResolveToken(ctx context.Context, req credential.TokenSpec) (*credential.TokenResult, error) {
	return &credential.TokenResult{Token: m.token, Scopes: m.scopes}, nil
}
