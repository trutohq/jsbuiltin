package jsbuiltin

import "testing"

// BenchmarkToLowerCase: the ASCII fast path (a header list, a domain) and the
// slow path a non-ASCII rune takes.
func BenchmarkToLowerCase(b *testing.B) {
	for _, c := range []struct{ name, in string }{
		{"ascii", "Authorization, Content-Type"},
		{"ascii-lower", "authorization,content-type"},
		{"non-ascii", "S14-İΣ.Example"},
	} {
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = ToLowerCase(c.in)
			}
		})
	}
}
