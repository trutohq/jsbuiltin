package jsbuiltin

import "strings"

// EncodeURIComponent is ECMAScript's `encodeURIComponent`: every UTF-8 byte of
// a code point outside `A-Z a-z 0-9 - _ . ! ~ * ' ( )` percent-encoded in
// upper case, and a lone surrogate a URIError ("URI malformed").
//
// The module's one implementation: /sso-connection's DNS-over-HTTPS name and
// the Sentinel forwarder's tenant, rule and stream segments use it.
func EncodeURIComponent(s string) (string, error) {
	if HasLoneSurrogate(s) {
		return "", &URIError{}
	}
	var b strings.Builder
	b.Grow(len(s))
	const hex = "0123456789ABCDEF"
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' ||
			strings.IndexByte("-_.!~*'()", c) >= 0 {
			b.WriteByte(c)
			continue
		}
		b.WriteByte('%')
		b.WriteByte(hex[c>>4])
		b.WriteByte(hex[c&15])
	}
	return b.String(), nil
}

// URIError is the URIError `encodeURIComponent` throws for a lone surrogate.
type URIError struct{}

func (*URIError) Error() string { return "URI malformed" }

// JSName is the class the request record's `logs` entry names.
func (*URIError) JSName() string { return "URIError" }

// EncodeURI is ECMAScript's encodeURI: every character but the unreserved
// marks and the URI's reserved delimiters (`;,/?:@&=+$#`) is UTF-8
// percent-encoded, and a lone surrogate is the URIError.
func EncodeURI(s string) (string, error) {
	if HasLoneSurrogate(s) {
		return "", &URIError{}
	}
	var b strings.Builder
	b.Grow(len(s))
	const hex = "0123456789ABCDEF"
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' ||
			strings.IndexByte("-_.!~*'();,/?:@&=+$#", c) >= 0 {
			b.WriteByte(c)
			continue
		}
		b.WriteByte('%')
		b.WriteByte(hex[c>>4])
		b.WriteByte(hex[c&15])
	}
	return b.String(), nil
}
