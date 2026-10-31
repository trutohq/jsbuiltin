package jsbuiltin

import (
	"math"
	"strings"
	"testing"
)

// TestParseIntPastFloat64IsInfinity pins parseInt(s, 10) on a digit run
// past float64's range to V8's answer (node 22, 2026-10-16): ±Infinity, not
// NaN. Batch B2 found it on /scim/v2's startIndex.
func TestParseIntPastFloat64IsInfinity(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want float64
	}{
		{"1" + strings.Repeat("0", 400), math.Inf(1)},
		{"-1" + strings.Repeat("0", 400), math.Inf(-1)},
		{strings.Repeat("9", 309), math.Inf(1)},
		{"1" + strings.Repeat("0", 308), 1e308},
		{" +" + strings.Repeat("7", 320) + "x", math.Inf(1)},
	} {
		got, ok := ParseInt(tc.in)
		if !ok || got != tc.want {
			t.Errorf("ParseInt(%d digits) = %v, %v; want %v", len(tc.in), got, ok, tc.want)
		}
	}
}
