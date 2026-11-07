package jsbuiltin

import (
	"sync"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// uppers hands out Upper casers, which are stateful and not safe to share.
var uppers = sync.Pool{New: func() any { c := cases.Upper(language.Und); return &c }}

// ToUpperCase is String.prototype.toUpperCase: the full, locale-independent
// mapping, where "ß" becomes "SS" and "ŉ" becomes "ʼN", which Go's per-rune
// strings.ToUpper does not do.
//
// Moved from truto-jsonata-go's jsstring.go (jsToUpperCase), 2026-10-31.
// Its mapping is x/text's, so it moves with x/text's Unicode version; V8's
// moves with ICU's. TestToUpperCaseIsV8s holds the two together on the
// runtime's goldens.
func ToUpperCase(s string) string { return ToUpperCaseOn(RuntimeV8, s) }

// ToUpperCaseOn is `String.prototype.toUpperCase()` on runtime rt.
// RuntimeBun is Bun 1.3.14, whose ICU predates the uppercase forms Unicode 16
// and 17 added: the 55 code points bunNoUpper names stay as they are there
// (testdata/bun/, TestBunContract).
func ToUpperCaseOn(rt Runtime, s string) string {
	if rt == RuntimeBun {
		for i, r := range s {
			if bunNoUpper(r) {
				n := len(string(r))
				return toUpperV8(s[:i]) + string(r) + ToUpperCaseOn(rt, s[i+n:])
			}
		}
	}
	return toUpperV8(s)
}

// bunNoUpper reports a code point whose uppercase form is new in Unicode 16
// or 17, which Bun 1.3.14 leaves as it is. Upper-casing has no context
// across code points, so the string is cut around it.
func bunNoUpper(r rune) bool {
	switch r {
	case 0x019B, 0x0264, 0x1C8A, 0xA7CD, 0xA7CF, 0xA7D3, 0xA7D5, 0xA7DB:
		return true
	}
	return (r >= 0x10D70 && r <= 0x10D85) || (r >= 0x16EBB && r <= 0x16ED3)
}

func toUpperV8(s string) string {
	ascii := true
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			ascii = false
			break
		}
	}
	if ascii {
		b := []byte(s)
		for i, c := range b {
			if c >= 'a' && c <= 'z' {
				b[i] = c - 32
			}
		}
		return string(b)
	}
	c := uppers.Get().(*cases.Caser)
	defer uppers.Put(c)
	return c.String(s)
}
