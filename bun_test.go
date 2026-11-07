package jsbuiltin_test

import (
	"encoding/json"
	"os"
	"testing"
	"unicode/utf8"

	"github.com/trutohq/jsbuiltin"
)

// TestBunContract holds RuntimeBun to Bun's own answers (testdata/bun/,
// oracle/bun.mjs on Bun 1.3.14). The lowercase list is complete, so every
// code point it does not name must lower to itself: that is the half a
// runtime-driven list alone cannot see, and where the 27 Unicode 16 code
// points were.
func TestBunContract(t *testing.T) {
	var rt struct{ Runtime, Version string }
	raw, err := os.ReadFile("testdata/bun/runtime.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &rt); err != nil || rt.Runtime != "bun" || rt.Version != "1.3.14" {
		t.Fatalf("testdata/bun is not Bun 1.3.14's: %s", raw)
	}
	cases := readJSONL[struct {
		Op  string `json:"op"`
		In  string `json:"in"`
		Out string `json:"out"`
		Err string `json:"err"`
	}](t, "testdata/bun/bun.jsonl")

	lowered, uppered := map[rune]bool{}, map[rune]bool{}
	n := 0
	for _, c := range cases {
		n++
		switch c.Op {
		case "lower", "lower-context":
			if c.Op == "lower" {
				r, _ := utf8.DecodeRuneInString(c.In)
				lowered[r] = true
			}
			if got := jsbuiltin.ToLowerCaseOn(jsbuiltin.RuntimeBun, c.In); got != c.Out {
				t.Errorf("ToLowerCaseOn(Bun, %q) = %q, Bun %q", c.In, got, c.Out)
			}
		case "upper", "upper-context":
			if c.Op == "upper" {
				r, _ := utf8.DecodeRuneInString(c.In)
				uppered[r] = true
			}
			if got := jsbuiltin.ToUpperCaseOn(jsbuiltin.RuntimeBun, c.In); got != c.Out {
				t.Errorf("ToUpperCaseOn(Bun, %q) = %q, Bun %q", c.In, got, c.Out)
			}
		case "uri-error":
			f := jsbuiltin.EncodeURIComponentOn
			if c.In == "encodeURI" {
				f = jsbuiltin.EncodeURIOn
			}
			_, err := f(jsbuiltin.RuntimeBun, "\xed\xa0\x80") // U+D800 in WTF-8
			if err == nil || "URIError: "+err.Error() != c.Err {
				t.Errorf("%s(Bun, lone surrogate) error %v, Bun %q", c.In, err, c.Err)
			}
		case "encodeURIComponent", "encodeURI":
			f := jsbuiltin.EncodeURIComponentOn
			if c.Op == "encodeURI" {
				f = jsbuiltin.EncodeURIOn
			}
			if got, err := f(jsbuiltin.RuntimeBun, c.In); err != nil || got != c.Out {
				t.Errorf("%s(Bun, %q) = %q, %v; Bun %q", c.Op, c.In, got, err, c.Out)
			}
		default:
			t.Fatalf("unknown op %q", c.Op)
		}
	}
	if len(lowered) < 1000 {
		t.Fatalf("thin Bun goldens: %d lowercase code points", len(lowered))
	}
	for r := rune(0); r <= utf8.MaxRune; r++ {
		if (r >= 0xd800 && r <= 0xdfff) || lowered[r] {
			continue
		}
		if s := string(r); jsbuiltin.ToLowerCaseOn(jsbuiltin.RuntimeBun, s) != s {
			t.Errorf("ToLowerCaseOn(Bun, U+%04X) changes it; Bun leaves it", r)
		}
	}
	if len(uppered) < 1000 {
		t.Fatalf("thin Bun goldens: %d uppercase code points", len(uppered))
	}
	for r := rune(0); r <= utf8.MaxRune; r++ {
		if (r >= 0xd800 && r <= 0xdfff) || uppered[r] {
			continue
		}
		if s := string(r); jsbuiltin.ToUpperCaseOn(jsbuiltin.RuntimeBun, s) != s {
			t.Errorf("ToUpperCaseOn(Bun, U+%04X) changes it; Bun leaves it", r)
		}
	}
	// A cut around a Bun-unmapped code point keeps the rest of the string
	// upper-cased, both sides of it.
	if got := jsbuiltin.ToUpperCaseOn(jsbuiltin.RuntimeBun, "a\u019bßb"); got != "A\u019bSSB" {
		t.Errorf("ToUpperCaseOn(Bun, a ƛ ß b) = %q", got)
	}
	t.Logf("%d Bun answers, %d lowercase code points", n, len(lowered))
}

// The V8 profile is unchanged by the Bun one: the plain functions still give
// V8's message and lower the Unicode 16 code points.
func TestTheV8ProfileIsThePlainFunctions(t *testing.T) {
	if _, err := jsbuiltin.EncodeURIComponent("\xed\xa0\x80"); err == nil || err.Error() != "URI malformed" {
		t.Errorf("EncodeURIComponent(lone surrogate) error %v, want V8's URI malformed", err)
	}
	if got := jsbuiltin.ToLowerCase("\U00010D50"); got != "\U00010D70" {
		t.Errorf("ToLowerCase(U+10D50) = %q, want U+10D70 (V8, Unicode 16)", got)
	}
	if got := jsbuiltin.ToLowerCaseOn(jsbuiltin.RuntimeBun, "\U00010D50"); got != "\U00010D50" {
		t.Errorf("ToLowerCaseOn(Bun, U+10D50) = %q, want it unchanged", got)
	}
}
