package jsbuiltin

import (
	"strings"
	"unicode/utf8"
)

// utf8BOM is EF BB BF, the byte-order mark as UTF-8.
const utf8BOM = "\xef\xbb\xbf"

// DecodeUTF8 is `new TextDecoder('utf-8').decode(bytes)` — the WHATWG
// Encoding Standard's UTF-8 decode with its two defaults, `fatal: false` and
// `ignoreBOM: false`. It is the one implementation of that call in this
// module: the keyset cursor (AbstractService.ts:236-238) and `parseRawBody`
// (src/common/arrayBufferToString.ts) both reach it.
//
// # Two rules, and Go's string(bytes) follows neither
//
//   - ONE leading BOM is removed. Only one, and only at the very start: a second
//     BOM, or one anywhere else, stays as U+FEFF. Measured under bun:
//     [EF BB BF 61] is "a", [EF BB BF EF BB BF 61] is "\uFEFFa". Without this a
//     BOM-prefixed JSON body is "Invalid JSON body" in Go and a 200 on the
//     reference (REVIEW-SLICE7 A-44), and a BOM-prefixed cursor parses to a key
//     the reference never sees.
//   - Malformed input is never an error: each MAXIMAL SUBPART of an ill-formed
//     sequence becomes ONE U+FFFD. string(bytes) keeps the invalid bytes, and the
//     obvious fix — one U+FFFD per utf8.RuneError — emits one per BYTE, so a
//     truncated four-byte sequence gives two replacement characters where the
//     reference gives one. [EF BB] (a truncated BOM) is one U+FFFD.
func DecodeUTF8(b []byte) string {
	if len(b) >= 3 && string(b[:3]) == utf8BOM {
		b = b[3:]
	}
	if utf8.Valid(b) {
		return string(b)
	}
	return replaceIllFormed(b)
}

// WellFormedUTF8 is the second of DecodeUTF8's two rules alone: each maximal
// subpart of an ill-formed sequence becomes one U+FFFD, and a leading BOM is
// KEPT. It is what `new TextEncoder().encode(str)` hashes when `str` was read
// from bytes — TextEncoder encodes every code unit, U+FEFF included, so a
// hash of a stored string must not strip one (`sha256` in
// src/common/hashString.ts, the /system-prompt route's `prompt_sha256`).
func WellFormedUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	return replaceIllFormed([]byte(s))
}

func replaceIllFormed(b []byte) string {
	var out strings.Builder
	out.Grow(len(b))
	for i := 0; i < len(b); {
		r, size := utf8.DecodeRune(b[i:])
		if r != utf8.RuneError || size > 1 {
			out.WriteRune(r)
			i += size
			continue
		}
		out.WriteRune(utf8.RuneError)
		i += maximalSubpart(b[i:])
	}
	return out.String()
}

// maximalSubpart returns the length of the maximal subpart of the ill-formed
// sequence at the front of b, per Unicode 3.9's recommended practice — which is
// what the Encoding Standard's UTF-8 decoder implements.
func maximalSubpart(b []byte) int {
	c := b[0]
	var n int
	lo, hi := byte(0x80), byte(0xBF)
	switch {
	case c >= 0xC2 && c <= 0xDF:
		n = 2
	case c == 0xE0:
		n, lo = 3, 0xA0
	case c >= 0xE1 && c <= 0xEC:
		n = 3
	case c == 0xED:
		n, hi = 3, 0x9F
	case c >= 0xEE && c <= 0xEF:
		n = 3
	case c == 0xF0:
		n, lo = 4, 0x90
	case c >= 0xF1 && c <= 0xF3:
		n = 4
	case c == 0xF4:
		n, hi = 4, 0x8F
	default:
		return 1
	}
	if len(b) < 2 || b[1] < lo || b[1] > hi {
		return 1
	}
	for i := 2; i < n; i++ {
		if i >= len(b) || b[i] < 0x80 || b[i] > 0xBF {
			return i
		}
	}
	return n
}
