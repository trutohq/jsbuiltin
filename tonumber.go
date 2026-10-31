package jsbuiltin

import (
	"math"
	"strconv"
	"strings"
)

// ToNumber is JavaScript's `Number(value)` for a STRING — ECMA-262's
// StringToNumber, which is NOT parseInt and differs from it on inputs a client
// really sends.
//
// `@truto/ginger`'s router adapter reaches for it where AbstractService reaches
// for parseInt (src/common/ginger-compat/legacyListAdapter.ts:143-152 versus
// src/common/AbstractService.ts:100), so the SAME query string means different
// things on the two contracts:
//
//	?limit=1e3    parseInt -> 1      Number -> 1000
//	?limit=0x2    parseInt -> 0      Number -> 2
//	?limit=12abc  parseInt -> 12     Number -> NaN
//	?limit="  2 " parseInt -> 2      Number -> 2
//	?limit=       (never reached: the caller tests trim() !== "" first)
//
// All four measured against the running reference on /group-mapping.
//
// ok is false for NaN. The empty string is 0 and NOT NaN, which is why the
// reference checks that `trim()` is not the empty string before calling it at
// all.
func ToNumber(s string) (float64, bool) {
	// The common case — VictoriaLogs' counts, a query's limit — is a run of
	// decimal digits short enough to be exact: read it without the trim and
	// the float parser. Fifteen digits stay below 2^53.
	if n := len(s); n > 0 && n <= 15 {
		v := 0.0
		for i := 0; i < n; i++ {
			c := s[i]
			if c < '0' || c > '9' {
				goto general
			}
			v = v*10 + float64(c-'0')
		}
		return v, true
	}
general:
	t := strings.TrimFunc(s, IsWhiteSpace)
	if t == "" {
		return 0, true
	}
	// A non-decimal literal takes no sign, so this is checked before the sign.
	if len(t) > 2 && t[0] == '0' {
		var base int
		switch t[1] {
		case 'x', 'X':
			base = 16
		case 'o', 'O':
			base = 8
		case 'b', 'B':
			base = 2
		}
		if base != 0 {
			n, err := strconv.ParseUint(t[2:], base, 64)
			if err != nil {
				// A digit outside the radix, or a value past uint64. The second
				// is a real number in JavaScript, which has no integer type at
				// all — reachable only from a limit with 20+ hex digits, and
				// falling back to NaN there is a divergence this records rather
				// than pretends away. See DIVERGENCES I-23.
				return 0, false
			}
			return float64(n), true
		}
	}
	body := t
	sign := 1.0
	if body[0] == '+' || body[0] == '-' {
		if body[0] == '-' {
			sign = -1
		}
		body = body[1:]
	}
	if body == "Infinity" {
		return sign * math.Inf(1), true
	}
	// strconv.ParseFloat accepts "inf", "nan", "0x1p-2" and underscores, none
	// of which StringToNumber does; the shape is checked first so those become
	// NaN as they should.
	if !isDecimalLiteral(body) {
		return 0, false
	}
	f, err := strconv.ParseFloat(body, 64)
	if err != nil {
		// Out of float64's range: JavaScript gives Infinity, and so does
		// ParseFloat's ErrRange result, which it returns alongside the error.
		if ne, ok := err.(*strconv.NumError); ok && ne.Err == strconv.ErrRange {
			return sign * f, true
		}
		return 0, false
	}
	return sign * f, true
}

// isDecimalLiteral is ECMA-262's StrUnsignedDecimalLiteral, minus Infinity:
// digits with an optional fraction and an optional exponent, where at least one
// of the integer and fraction parts must have a digit.
func isDecimalLiteral(s string) bool {
	i, digits := 0, 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i, digits = i+1, digits+1
	}
	if i < len(s) && s[i] == '.' {
		i++
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i, digits = i+1, digits+1
		}
	}
	if digits == 0 {
		return false
	}
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		i++
		if i < len(s) && (s[i] == '+' || s[i] == '-') {
			i++
		}
		expDigits := 0
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i, expDigits = i+1, expDigits+1
		}
		if expDigits == 0 {
			return false
		}
	}
	return i == len(s)
}
