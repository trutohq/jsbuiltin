package jsbuiltin_test

import (
	"strings"
	"testing"

	"github.com/trutohq/jsbuiltin"
)

// TestTrimMatchesLodash pins ECMAScript's `\s` against lodash's own `trim`.
//
// Every pair was MEASURED, and the provenance is stated precisely because the
// first version of this comment got it wrong in the exact way H-79 warns about:
// it said "`bun /tmp/trimtest.mjs`, lodash-es 4.17 from this repository's
// node_modules", and a probe at `/tmp` does not resolve this repository's
// node_modules at all (REVIEW-SLICE6 R-8). It got away with it — lodash-es
// resolves to the same copy either way — but the sentence was self-refuting.
//
// Re-measured 2026-09-22 from a probe inside the worktree importing
// `./node_modules/lodash-es/lodash.js` by relative path, against **lodash-es
// 4.18.1** (`internal/rig`'s TestReferenceLibrariesResolveAsMeasured pins that
// version). Every pair below reproduces. The two that matter are the two Go
// disagrees about:
//
//	U+FEFF  ZWNBSP  ECMAScript trims it, unicode.IsSpace does not
//	U+0085  NEL     unicode.IsSpace trims it, ECMAScript does not
//
// `TeamService.invite` decides whether an environment id is blank with
// `compact(map(environment, trim))` (TeamService.ts:203), so a disagreement is
// the difference between a 400 and a membership row pointing at an environment
// that does not exist.
func TestTrimMatchesLodash(t *testing.T) {
	// Built from code points so the source file carries no invisible bytes —
	// a literal U+FEFF in a Go file is an "illegal byte order mark", and a
	// literal U+0085 is invisible in every diff tool.
	nel := string(rune(0x0085))
	bom := string(rune(0xfeff))
	cases := []struct{ name, in, want string }{
		{"tab-and-newline", "\t a \n", "a"},
		{"spaces", " x ", "x"},
		{"zwnbsp-is-trimmed", bom + "y" + bom, "y"},
		{"nothing-to-trim", "z", "z"},
		{"nbsp", " w ", "w"},
		{"ideographic-space", "　v　", "v"},
		{"nel-is-NOT-trimmed", nel + "n" + nel, nel + "n" + nel},
		{"line-and-paragraph-separator", " s ", "s"},
		{"all-whitespace", "   ", ""},
		{"empty", "", ""},
		{"plain", "ok", "ok"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := jsbuiltin.Trim(tc.in); got != tc.want {
				t.Fatalf("Trim(%q) = %q, want %q (measured against lodash)", tc.in, got, tc.want)
			}
		})
	}
}

// TestTrimIsNotTrimSpace makes the reason this function exists a test rather
// than a comment: if the two ever agreed, the helper could go, and if someone
// replaces it with strings.TrimSpace this fails.
func TestTrimIsNotTrimSpace(t *testing.T) {
	for _, tc := range []struct{ name, in string }{
		{"NEL", string(rune(0x0085)) + "a" + string(rune(0x0085))},
		{"ZWNBSP", string(rune(0xfeff)) + "a" + string(rune(0xfeff))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if jsbuiltin.Trim(tc.in) == strings.TrimSpace(tc.in) {
				t.Fatalf("Trim and strings.TrimSpace agree on %q; one of them has changed and "+
					"the lodash measurement in TestTrimMatchesLodash is the authority", tc.in)
			}
		})
	}
}
