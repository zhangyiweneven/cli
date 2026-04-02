package credential

import "testing"

func TestTokenTypeString(t *testing.T) {
	tests := []struct {
		tt   TokenType
		want string
	}{
		{TokenTypeUAT, "uat"},
		{TokenTypeTAT, "tat"},
		{TokenType(99), "unknown"},
	}
	for _, tc := range tests {
		if got := tc.tt.String(); got != tc.want {
			t.Errorf("TokenType(%d).String() = %q, want %q", tc.tt, got, tc.want)
		}
	}
}

func TestParseTokenType(t *testing.T) {
	tests := []struct {
		s    string
		want TokenType
		ok   bool
	}{
		{"uat", TokenTypeUAT, true},
		{"tat", TokenTypeTAT, true},
		{"UAT", TokenTypeUAT, true},
		{"bad", 0, false},
	}
	for _, tc := range tests {
		got, ok := ParseTokenType(tc.s)
		if ok != tc.ok || (ok && got != tc.want) {
			t.Errorf("ParseTokenType(%q) = (%v, %v), want (%v, %v)", tc.s, got, ok, tc.want, tc.ok)
		}
	}
}
