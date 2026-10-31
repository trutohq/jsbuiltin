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
func ToUpperCase(s string) string {
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
