package jsbuiltin

import (
	"strings"
	"testing"
)

// TestUTF16PrefixIsJavaScriptsSlice pins `s.slice(0, n)` on the shapes that
// separate code units from bytes and runes. The trusted-IP corpus
// (internal/clientip, measured on workerd) is what pins it against the
// reference; this is the unit half, so a regression names this function.
func TestUTF16PrefixIsJavaScriptsSlice(t *testing.T) {
	smile := "\U0001F600" // two units, four bytes
	hi := UTF16Unit(0xD83D)
	for _, tc := range []struct {
		name string
		in   string
		n    int
		want string
	}{
		{"ascii under", "abc", 64, "abc"},
		{"ascii cut", strings.Repeat("a", 5000), 64, strings.Repeat("a", 64)},
		{"zero", "abc", 0, ""},
		{"negative", "abc", -1, ""},
		{"two-byte runes count one unit each", strings.Repeat("é", 70), 64, strings.Repeat("é", 64)},
		{"three-byte runes count one unit each", strings.Repeat("日", 70), 64, strings.Repeat("日", 64)},
		{"astral pair fits", "a" + smile, 3, "a" + smile},
		{"cut between the halves keeps the high one", "a" + smile, 2, "a" + hi},
		{"cut before the pair", "ab" + smile, 2, "ab"},
		{"lone surrogate counts one unit", hi + "xyz", 2, hi + "x"},
		{"exactly n units", strings.Repeat(smile, 32), 64, strings.Repeat(smile, 32)},
		{"one over", strings.Repeat(smile, 32) + "a", 64, strings.Repeat(smile, 32)},
		{"odd cut through pairs", strings.Repeat(smile, 40), 63, strings.Repeat(smile, 31) + hi},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := UTF16Prefix(tc.in, tc.n)
			if got != tc.want {
				t.Fatalf("UTF16Prefix(%q, %d) = %q, want %q", tc.in, tc.n, got, tc.want)
			}
			if tc.n > 0 && UTF16Len(got) > tc.n {
				t.Fatalf("the prefix is %d units, over the bound %d", UTF16Len(got), tc.n)
			}
		})
	}
}

// TestCompareUTF16IsJavaScriptsStringOrder pins `a < b` on strings: UTF-16
// code unit order, which byte order gets wrong only for a supplementary
// character against U+E000…U+FFFF.
func TestCompareUTF16IsJavaScriptsStringOrder(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want int
	}{
		{"", "", 0}, {"", "a", -1}, {"a", "", 1}, {"abc", "abd", -1}, {"abc", "ab", 1}, {"ab", "abc", -1},
		{"Z", "a", -1}, {"é", "z", 1}, {"a😀", "a😀", 0},
		{"￿", "😀", 1}, {"😀", "￿", -1}, {"", "😀", 1}, {"퟿", "😀", -1},
		{"😀", "😁", -1}, {"x￿y", "x😀", 1}, {"日本", "日本語", -1}, {"日本a", "日本b", -1},
		// A lone surrogate is one code unit, below U+E000 (A-79's WTF-8).
		{"\xed\xa0\x80", "\ue000", -1}, {"\xed\xa0\x80", "\ud7ff", 1},
	} {
		if got := CompareUTF16(c.a, c.b); sign(got) != c.want {
			t.Errorf("CompareUTF16(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	}
	return 0
}

// compareByUnits is the definition: UTF16Units compared unit by unit.
func compareByUnits(a, b string) int {
	ua, ub := UTF16Units(a), UTF16Units(b)
	for i := 0; i < len(ua) && i < len(ub); i++ {
		if ua[i] != ub[i] {
			return sign(int(ua[i]) - int(ub[i]))
		}
	}
	return sign(len(ua) - len(ub))
}

// FuzzCompareUTF16 holds CompareUTF16 to its definition on any bytes:
// invalid UTF-8, a WTF-8 lone surrogate and a prefix ending inside a
// sequence included.
func FuzzCompareUTF16(f *testing.F) {
	for _, s := range [][2]string{{"a", "b"}, {"\xe2\x82", "\xe2\x82\xac"}, {"\xe2\x82x", "\xe2\x82"}, {"￿", "😀"},
		{"ab\xff", "ab"}, {"\xed\xa0\x80", "\xed\xa0\x81"}} {
		f.Add(s[0], s[1])
	}
	f.Fuzz(func(t *testing.T, a, b string) {
		if got, want := sign(CompareUTF16(a, b)), compareByUnits(a, b); got != want {
			t.Fatalf("CompareUTF16(%q, %q) = %d, by units %d", a, b, got, want)
		}
	})
}
