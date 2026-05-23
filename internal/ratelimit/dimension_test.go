package ratelimit

import "testing"

func TestDimension_IsTPM(t *testing.T) {
	tests := []struct {
		dim  Dimension
		want bool
	}{
		{DimUserRPM, false},
		{DimUserTPM, true},
		{DimTokenRPM, false},
		{DimTokenTPM, true},
		{DimIPRPM, false},
		{DimModelRPM, false},
		{DimModelTPM, true},
	}
	for _, tc := range tests {
		if got := tc.dim.IsTPM(); got != tc.want {
			t.Errorf("dim=%s IsTPM=%v want %v", tc.dim, got, tc.want)
		}
	}
}

func TestDimension_HeaderSuffix(t *testing.T) {
	tests := []struct {
		dim  Dimension
		want string
	}{
		{DimUserRPM, "User-RPM"},
		{DimUserTPM, "User-TPM"},
		{DimTokenRPM, "Token-RPM"},
		{DimTokenTPM, "Token-TPM"},
		{DimIPRPM, "IP-RPM"},
		{DimModelRPM, "Model-RPM"},
		{DimModelTPM, "Model-TPM"},
	}
	for _, tc := range tests {
		if got := tc.dim.HeaderSuffix(); got != tc.want {
			t.Errorf("dim=%s HeaderSuffix=%q want %q", tc.dim, got, tc.want)
		}
	}
}

func TestDimension_HeaderSuffix_Unknown(t *testing.T) {
	// 未知 Dimension 转回 string 不致 panic
	if got := Dimension("custom").HeaderSuffix(); got != "Custom" {
		t.Errorf("custom HeaderSuffix=%q want Custom", got)
	}
}
