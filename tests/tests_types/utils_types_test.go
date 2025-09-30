package testtypes

import (
	"testing"

	types "github.com/kubex-ecosystem/gobe/internal/contracts/types"
)

func TestIsShellSpecialVar(t *testing.T) {
	cases := []struct {
		c    byte
		want bool
	}{
		{'*', true}, {'#', true}, {'$', true}, {'A', false}, {'z', false},
	}
	for _, tc := range cases {
		if got := types.IsShellSpecialVar(tc.c); got != tc.want {
			t.Fatalf("IsShellSpecialVar(%q) = %v, want %v", tc.c, got, tc.want)
		}
	}
}

func TestIsAlphaNum(t *testing.T) {
	cases := []struct {
		c    byte
		want bool
	}{
		{'_', true}, {'0', true}, {'9', true}, {'a', true}, {'Z', true}, {'-', false}, {' ', false},
	}
	for _, tc := range cases {
		if got := types.IsAlphaNum(tc.c); got != tc.want {
			t.Fatalf("IsAlphaNum(%q) = %v, want %v", tc.c, got, tc.want)
		}
	}
}
