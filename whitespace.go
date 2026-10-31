package jsbuiltin

import (
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"
)

// IsWhiteSpace is ECMA-262's StrWhiteSpaceChar (WhiteSpace and
// LineTerminator): the set String.prototype.trim removes, `\s` matches,
// parseInt skips and Number() trims.
//
// It is NOT unicode.IsSpace: that includes U+0085 (NEL), which JavaScript
// does not, and excludes U+FEFF, which JavaScript includes. The table is
// written out rather than read from unicode.Zs so it does not move with the
// Go toolchain's Unicode version; TestWhiteSpaceIsV8s holds it to the
// runtime for every code point.
//
// This consolidates two copies that had the same set: gotruto's jsonx
// IsWhiteSpace (unicode.Zs plus the rest) and its qsx isJSWhitespace (the
// table).
func IsWhiteSpace(r rune) bool {
	switch r {
	case '\t', '\n', '\v', '\f', '\r', ' ', 0x00A0, 0xFEFF,
		0x1680, 0x2000, 0x2001, 0x2002, 0x2003, 0x2004, 0x2005, 0x2006,
		0x2007, 0x2008, 0x2009, 0x200A, 0x2028, 0x2029, 0x202F, 0x205F, 0x3000:
		return true
	}
	return false
}

// Trim is String.prototype.trim.
//
// It matters wherever a value is tested for blankness: TeamService.invite
// decides whether an environment id is blank with
// compact(map(environment, trim)) (truto src/team/TeamService.ts:203), and a
// value one side trims to "" and the other does not is a 400 on one side and
// a membership row on the other.
func Trim(s string) string { return strings.TrimFunc(s, IsWhiteSpace) }

// ParseInt is JavaScript's parseInt(s, 10).
//
// It skips leading whitespace, accepts a sign, reads decimal digits and STOPS
// at the first non-digit, so ParseInt("12abc") is 12 where strconv.Atoi is an
// error, ParseInt("1e3") is 1, and ParseInt("") is NaN. ok is false only
// where JavaScript returns NaN, i.e. no digits at all.
//
// The digits are handed to ParseFloat rather than accumulated, so a long digit
// string rounds to the same float64 JavaScript produces, and a run past
// float64's range is ±Infinity, as parseInt("1" + "0".repeat(400)) is.
func ParseInt(s string) (float64, bool) {
	i := 0
	for i < len(s) {
		r, size := utf8.DecodeRuneInString(s[i:])
		if !IsWhiteSpace(r) {
			break
		}
		i += size
	}
	start := i
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		i++
	}
	digits := i
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == digits {
		return 0, false
	}
	f, err := strconv.ParseFloat(s[start:i], 64)
	if err != nil && !errors.Is(err, strconv.ErrRange) {
		return 0, false
	}
	return f, true
}
