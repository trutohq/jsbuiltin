package jsbuiltin

import (
	"encoding/base64"
	"errors"
)

// The two ways atob refuses a string. Both are a DOMException named
// "InvalidCharacterError" in JavaScript; the runtimes word them differently,
// so the message a caller prints is the caller's (Node's are these; workerd's
// is one sentence for both, gotruto's cursor.WorkerdAtobMessage).
var (
	// ErrInvalidCharacter is a byte outside the alphabet, an inner "=", more
	// than two "=", or padding that does not complete a group of four.
	ErrInvalidCharacter = errors.New("Invalid character")

	// ErrNotCorrectlyEncoded is a single character left over after the last
	// group of four.
	ErrNotCorrectlyEncoded = errors.New("The string to be decoded is not correctly encoded.")
)

// Atob is the HTML Standard's forgiving-base64 decode, which atob runs:
// ASCII whitespace is dropped, one or two trailing "=" are allowed when they
// complete a group of four, input of length 2 or 3 mod 4 needs no "=", and
// leftover bits are discarded rather than checked. base64.StdEncoding refuses
// most of that, and a stricter decoder refuses cursors the reference accepts.
//
// Which error a bad input gets follows Node (simdutf's forgiving decode,
// which tells the two apart): more than two "=" or any byte outside the
// alphabet is [ErrInvalidCharacter], checked before a lone leftover character
// ([ErrNotCorrectlyEncoded]), checked before padding that does not complete a
// group ([ErrInvalidCharacter] again). Whether a string is accepted, and the
// bytes, do not depend on the runtime.
//
// s is the string's UTF-8; a code point past U+007F is outside the alphabet
// either way.
//
// Consolidates truto-jsonata-go's fn_crypto.go atob (this code), gotruto's
// cursor.atob, analytics.atob and saml.atob (all accept and refuse the same
// strings; only the module's told the two errors apart).
func Atob(s string) ([]byte, error) {
	data := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '\t', '\n', '\f', '\r', ' ':
		default:
			data = append(data, c)
		}
	}
	pad := 0
	for n := len(data); pad < n && data[n-1-pad] == '='; pad++ {
	}
	if pad > 2 {
		return nil, ErrInvalidCharacter
	}
	content := data[:len(data)-pad]
	out := make([]byte, 0, len(content)*3/4)
	var buf uint32
	bits := 0
	for _, c := range content {
		var v byte
		switch {
		case c >= 'A' && c <= 'Z':
			v = c - 'A'
		case c >= 'a' && c <= 'z':
			v = c - 'a' + 26
		case c >= '0' && c <= '9':
			v = c - '0' + 52
		case c == '+':
			v = 62
		case c == '/':
			v = 63
		default:
			return nil, ErrInvalidCharacter
		}
		buf = buf<<6 | uint32(v)
		bits += 6
		if bits >= 8 {
			bits -= 8
			out = append(out, byte(buf>>bits))
		}
	}
	if len(content)%4 == 1 {
		return nil, ErrNotCorrectlyEncoded
	}
	if pad > 0 && len(data)%4 != 0 {
		return nil, ErrInvalidCharacter
	}
	return out, nil
}

// ErrBtoaRange is btoa's refusal of a code unit above 0xFF.
var ErrBtoaRange = errors.New("Invalid character")

// Btoa is btoa(s) for a JavaScript string held as UTF-8 (a lone surrogate as
// WTF-8): each code unit is one byte, so "é" (U+00E9) is the byte 0xE9 and
// encodes as "6Q==", not as its UTF-8 "w6k=". A code unit above 0xFF is
// [ErrBtoaRange]; a code point above U+FFFF is a surrogate pair whose halves
// are both above 0xFF, so testing the rune is testing the units.
func Btoa(s string) (string, error) {
	b := make([]byte, 0, len(s))
	for _, r := range s {
		if r > 0xFF {
			return "", ErrBtoaRange
		}
		b = append(b, byte(r))
	}
	return BtoaBytes(b), nil
}

// BtoaBytes is btoa over a binary string whose code units are these bytes:
// standard alphabet, padded.
func BtoaBytes(b []byte) string { return base64.StdEncoding.EncodeToString(b) }
