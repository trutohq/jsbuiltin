package jsbuiltin

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// ToLowerCase is `String.prototype.toLowerCase()` as main's workerd runs it
// (V8 over ICU, Unicode 16), for any string.
//
// Go's strings.ToLower is a per-rune SIMPLE mapping on Go's Unicode tables,
// and differs from V8's in three ways, each measured over all of Unicode on
// 2026-09-25 (I-49, REVIEW-SLICE10):
//
//   - U+0130 LATIN CAPITAL LETTER I WITH DOT ABOVE lowers to TWO code points,
//     "i" and U+0307, in V8 (SpecialCasing.txt's unconditional mapping), and
//     to "i" in Go;
//   - U+03A3 GREEK CAPITAL LETTER SIGMA lowers to ς (U+03C2) where it ends a
//     word — Final_Sigma, ECMA-262 §22.1.3.28 — and to σ everywhere else; Go
//     always gives σ;
//   - the 28 code points Unicode 17 gave a lowercase form (U+A7CE, U+A7D2,
//     U+A7D4, U+16EA0…U+16EB8): Go 1.27's tables lower them, workerd's
//     Unicode 16 leaves them alone.
//
// Each is reproduced here. One implementation for the port: the trusted
// client-IP header (internal/clientip) and the SSO domain (slice 14) both
// call it.
func ToLowerCase(s string) string {
	upper := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 0x80 {
			upper = false
			goto slow
		}
		if c >= 'A' && c <= 'Z' {
			upper = true
		}
	}
	// All ASCII: the stdlib is exact, and nothing to do without a capital.
	// The WHOLE string has to be scanned first — a capital before a non-ASCII
	// rune once took this path and lowered U+0130 the stdlib's way.
	if upper {
		return strings.ToLower(s)
	}
	return s
slow:
	var b strings.Builder
	b.Grow(len(s) + 4)
	for i, r := range s {
		switch {
		case r == 0x0130:
			b.WriteString("i\u0307")
		case r == 0x03A3:
			if finalSigma(s, i) {
				b.WriteRune(0x03C2)
			} else {
				b.WriteRune(0x03C3)
			}
		case unicode17Lowercase(r):
			b.WriteRune(r)
		default:
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

// unicode17Lowercase reports a code point whose lowercase mapping is new in
// Unicode 17, which workerd's V8 does not have yet.
func unicode17Lowercase(r rune) bool {
	return r == 0xA7CE || r == 0xA7D2 || r == 0xA7D4 || (r >= 0x16EA0 && r <= 0x16EB8)
}

// finalSigma is Unicode's Final_Sigma condition for the Σ at byte offset i:
// preceded by a cased letter (skipping case-ignorable characters) and not
// followed by one (likewise).
func finalSigma(s string, i int) bool {
	before := false
	for j := i; j > 0; {
		r, size := utf8.DecodeLastRuneInString(s[:j])
		j -= size
		if caseIgnorable(r) {
			continue
		}
		before = cased(r)
		break
	}
	if !before {
		return false
	}
	for j := i + utf8.RuneLen(0x03A3); j < len(s); {
		r, size := utf8.DecodeRuneInString(s[j:])
		j += size
		if caseIgnorable(r) {
			continue
		}
		return !cased(r)
	}
	return true
}

// cased is Unicode's Cased property, as far as the tables Go ships allow:
// Lowercase, Uppercase, or Titlecase letters, plus the Other_Lowercase and
// Other_Uppercase sets.
func cased(r rune) bool {
	return unicode.IsUpper(r) || unicode.IsLower(r) || unicode.IsTitle(r) ||
		unicode.Is(unicode.Other_Lowercase, r) || unicode.Is(unicode.Other_Uppercase, r)
}

// caseIgnorable is Unicode's Case_Ignorable property: Mn, Me, Cf, Lm, Sk, and
// the Word_Break MidLetter / MidNumLet / Single_Quote characters.
func caseIgnorable(r rune) bool {
	switch r {
	case '\'', '.', ':', '^', '`', 0x00B7, 0x0387, 0x05F4, 0x2018, 0x2019, 0x2024, 0x2027, 0xFE13,
		0xFE52, 0xFE55, 0xFF07, 0xFF0E, 0xFF1A:
		return true
	}
	return unicode.In(r, unicode.Mn, unicode.Me, unicode.Cf, unicode.Lm, unicode.Sk)
}
