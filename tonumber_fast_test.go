package jsbuiltin

import (
	"strconv"
	"testing"
)

// FuzzToNumberDigitFastPath holds ToNumber's run-of-digits shortcut to the
// general path it skips: the same value for every string, digits or not.
func FuzzToNumberDigitFastPath(f *testing.F) {
	for _, s := range []string{"0", "007", "123456789012345", "1234567890123456", "", " 1", "1 ", "+1", "1e3", "٣"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		got, ok := ToNumber(s)
		// The general path, as ToNumber ran before the shortcut: force it
		// with a trailing space it trims away.
		want, wantOK := ToNumber(s + " ")
		if len(s) > 0 && s[len(s)-1] == ' ' || s == "" {
			return
		}
		if ok != wantOK || ok && got != want {
			t.Fatalf("ToNumber(%q) = %v, %v; the general path %v, %v", s, got, ok, want, wantOK)
		}
		if allDigits(s) && len(s) <= 15 {
			if p, err := strconv.ParseFloat(s, 64); err != nil || p != got {
				t.Fatalf("ToNumber(%q) = %v, ParseFloat %v", s, got, p)
			}
		}
	})
}

func allDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return s != ""
}
