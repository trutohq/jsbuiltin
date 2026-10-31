package jsbuiltin_test

import (
	"testing"

	"github.com/trutohq/jsbuiltin"
)

func TestParseInt(t *testing.T) {
	tests := []struct {
		in   string
		want float64
		ok   bool
	}{
		{"12", 12, true},
		{"12abc", 12, true},
		{"1e3", 1, true},
		{"", 0, false},
		{"abc", 0, false},
		{"  7  ", 7, true},
		{"+5", 5, true},
		{"-5", -5, true},
		{"0012", 12, true},
		{"9007199254740993", 9007199254740992, true},
	}
	for _, tc := range tests {
		got, ok := jsbuiltin.ParseInt(tc.in)
		if ok != tc.ok || (ok && got != tc.want) {
			t.Errorf("ParseInt(%q) = %v, %v; want %v, %v", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}
