package jsbuiltin

import (
	"unicode/utf16"
	"unicode/utf8"
)

// UTF16Len is a JavaScript string's `length` for a Go string: UTF-16 code
// units, not bytes and not runes — an astral character counts twice.
//
// The one implementation in the module: superstruct's size(), the tenant name
// refinement and the transactional mail's rule widths all measure with it. It
// agrees with [UTF16Units] on every string, a lone surrogate included.
func UTF16Len(s string) int {
	n := 0
	for i := 0; i < len(s); {
		if _, ok := wtf8Surrogate(s[i:]); ok {
			n++
			i += 3
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		n++
		if r > 0xFFFF {
			n++
		}
		i += size
	}
	return n
}

// UTF16Units is a Go string as the JavaScript string it stands for: its UTF-16
// code units. An astral rune is a pair; a LONE surrogate, which the port holds
// as WTF-8 (ED A0..BF 80..BF — store's lodash get answers one for an index into
// half of an astral character, A-79), is one unit, as it is in JavaScript; any
// other invalid byte is U+FFFD, as a range over the string has it.
func UTF16Units(s string) []uint16 {
	units := make([]uint16, 0, len(s))
	for i := 0; i < len(s); {
		if u, ok := wtf8Surrogate(s[i:]); ok {
			units = append(units, u)
			i += 3
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if r > 0xFFFF {
			hi, lo := utf16.EncodeRune(r)
			units = append(units, uint16(hi), uint16(lo))
		} else {
			units = append(units, uint16(r))
		}
		i += size
	}
	return units
}

// UTF16Prefix is `s.slice(0, n)` for a non-negative n: the first n UTF-16
// code units of s, as the Go string for them.
//
// A cut through an astral character leaves its HIGH surrogate alone at the
// end, which JavaScript keeps; here it is that surrogate in WTF-8
// ([UTF16Unit]), so two strings that differ only after the cut still come out
// equal and two that differ before it do not — the property a length-capped
// map key needs (getTrustedClientIP's `v.slice(0, 64)`). A string that is
// already short enough is returned as it is.
func UTF16Prefix(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if len(s) <= n {
		// Every UTF-16 unit is at least one byte in WTF-8, so a string of at
		// most n bytes has at most n units.
		return s
	}
	units := 0
	for i := 0; i < len(s); {
		if _, ok := wtf8Surrogate(s[i:]); ok {
			if units+1 > n {
				return s[:i]
			}
			units++
			i += 3
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		w := 1
		if r > 0xFFFF {
			w = 2
		}
		if units+w > n {
			if w == 2 && units+1 == n {
				// The cut falls between the halves of a pair: JavaScript keeps
				// the high one.
				hi, _ := utf16.EncodeRune(r)
				return s[:i] + UTF16Unit(uint16(hi))
			}
			return s[:i]
		}
		units += w
		i += size
	}
	return s
}

// UTF16Unit is one code unit as a Go string — `str[i]` in JavaScript. A
// surrogate is WTF-8, so a lone one survives as itself rather than as U+FFFD,
// which is a different character with different bytes (A-79).
func UTF16Unit(u uint16) string {
	if u >= 0xD800 && u <= 0xDFFF {
		return string([]byte{0xED, 0xA0 | byte(u>>6&0x1F), 0x80 | byte(u&0x3F)})
	}
	return string(rune(u))
}

// HasLoneSurrogate reports whether s holds a surrogate in WTF-8, which valid
// UTF-8 never contains.
func HasLoneSurrogate(s string) bool {
	for i := 0; i+2 < len(s); i++ {
		if _, ok := wtf8Surrogate(s[i:]); ok {
			return true
		}
	}
	return false
}

// wtf8Surrogate reads a surrogate code point in WTF-8 at the start of s.
func wtf8Surrogate(s string) (uint16, bool) {
	if len(s) >= 3 && s[0] == 0xED && s[1] >= 0xA0 && s[1] <= 0xBF && s[2] >= 0x80 && s[2] <= 0xBF {
		return 0xD000 | uint16(s[1]&0x3F)<<6 | uint16(s[2]&0x3F), true
	}
	return 0, false
}

// UTF16String is the inverse of [UTF16Units]: code units back to the Go string
// the port holds for them. A high surrogate followed by a low one is the
// astral character they encode; a surrogate on its own stays itself, in WTF-8
// ([UTF16Unit]), so UTF16String(UTF16Units(s)) == s for every s UTF16Units
// accepts without replacing a byte.
func UTF16String(units []uint16) string {
	b := make([]byte, 0, len(units))
	for i := 0; i < len(units); i++ {
		u := units[i]
		switch {
		case u >= 0xD800 && u <= 0xDBFF && i+1 < len(units) && units[i+1] >= 0xDC00 && units[i+1] <= 0xDFFF:
			b = utf8.AppendRune(b, utf16.DecodeRune(rune(u), rune(units[i+1])))
			i++
		case u >= 0xD800 && u <= 0xDFFF:
			b = append(b, UTF16Unit(u)...)
		default:
			b = utf8.AppendRune(b, rune(u))
		}
	}
	return string(b)
}

// CompareUTF16 is JavaScript's `a < b` / `a > b` on strings: UTF-16 code
// unit order over the strings [UTF16Units] reads — an astral rune is its two
// surrogates, a lone surrogate held as WTF-8 is that one unit, any other
// invalid byte is U+FFFD. It differs from Go's byte order only where a
// surrogate (0xD800-0xDFFF) meets a code point from U+E000 to U+FFFF.
//
// The module's one string order: billing's JS value comparison and
// xmldsig's attribute sort use it too (REVIEW-BATCH-C2 A-328).
func CompareUTF16(a, b string) int {
	// The common prefix is the same units on both sides. Where the first
	// differing bytes are both ASCII, no multi-byte sequence can straddle
	// them, so they decide; where one string is a prefix of the other and the
	// longer goes on with an ASCII byte, the shorter is less. Anything else
	// takes the unit walk.
	n := min(len(a), len(b))
	i := 0
	for i < n && a[i] == b[i] {
		i++
	}
	switch {
	case i < n && a[i] < utf8.RuneSelf && b[i] < utf8.RuneSelf:
		if a[i] < b[i] {
			return -1
		}
		return 1
	case i == n && len(a) == len(b):
		return 0
	case i == n && len(a) < len(b) && b[n] < utf8.RuneSelf:
		return -1
	case i == n && len(b) < len(a) && a[n] < utf8.RuneSelf:
		return 1
	}
	ua, ub := unitReader{s: a, low: -1}, unitReader{s: b, low: -1}
	for {
		x, okA := ua.next()
		y, okB := ub.next()
		switch {
		case !okA && !okB:
			return 0
		case !okA:
			return -1
		case !okB:
			return 1
		case x != y:
			if x < y {
				return -1
			}
			return 1
		}
	}
}

// unitReader yields a string's UTF-16 code units, as UTF16Units does,
// without building the slice.
type unitReader struct {
	s   string
	low rune
}

func (u *unitReader) next() (rune, bool) {
	if u.low >= 0 {
		r := u.low
		u.low = -1
		return r, true
	}
	if u.s == "" {
		return 0, false
	}
	if v, ok := wtf8Surrogate(u.s); ok {
		u.s = u.s[3:]
		return rune(v), true
	}
	r, size := utf8.DecodeRuneInString(u.s)
	u.s = u.s[size:]
	if r > 0xFFFF {
		hi, lo := utf16.EncodeRune(r)
		u.low = lo
		return hi, true
	}
	return r, true
}
