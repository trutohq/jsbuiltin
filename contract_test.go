package jsbuiltin_test

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/trutohq/jsbuiltin"
)

// The contract: every answer in testdata/ was produced by executing the
// runtime (oracle/generate.mjs; testdata/runtime.json says which), and each
// must be the package's. The watch layer (TestWatch, `make watch`) runs the
// same comparison over goldens another runtime produced and reports, without
// failing, what moved.

type builtinCase struct {
	Op  string          `json:"op"`
	In  string          `json:"in"`
	Out json.RawMessage `json:"out"`
	Err string          `json:"err"`
}

type numberCase struct {
	Bits string `json:"bits"`
	Str  string `json:"str"`
}

func readJSONL[T any](t *testing.T, path string) []T {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var out []T
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<24)
	for sc.Scan() {
		var v T
		if err := json.Unmarshal(sc.Bytes(), &v); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		out = append(out, v)
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

func numStr(f float64, ok bool) string {
	if !ok {
		return "NaN"
	}
	if f == 0 && math.Signbit(f) {
		return "-0"
	}
	return jsbuiltin.FormatNumber(f)
}

// compareGoldens returns one line per answer in dir that the package does
// not give, and how many answers it compared.
func compareGoldens(t *testing.T, dir string) (diffs []string, n int) {
	t.Helper()
	for _, g := range readJSONL[numberCase](t, filepath.Join(dir, "numbers.jsonl")) {
		bits, err := strconv.ParseUint(g.Bits, 16, 64)
		if err != nil {
			t.Fatal(err)
		}
		n++
		if got := jsbuiltin.FormatNumber(math.Float64frombits(bits)); got != g.Str {
			diffs = append(diffs, fmt.Sprintf("Number::toString(0x%s): %q, runtime %q", g.Bits, got, g.Str))
		}
	}
	for _, c := range readJSONL[builtinCase](t, filepath.Join(dir, "builtins.jsonl")) {
		n++
		var want string
		if c.Err == "" {
			if c.Op == "whitespace" || c.Op == "trim-whitespace" {
				var cps []rune
				if err := json.Unmarshal(c.Out, &cps); err != nil {
					t.Fatal(err)
				}
				var got []rune
				for r := rune(0); r <= 0x10FFFF; r++ {
					if r >= 0xD800 && r <= 0xDFFF {
						continue
					}
					if c.Op == "whitespace" && jsbuiltin.IsWhiteSpace(r) ||
						c.Op == "trim-whitespace" && jsbuiltin.Trim(string(r)) == "" {
						got = append(got, r)
					}
				}
				if !slices.Equal(got, cps) {
					diffs = append(diffs, fmt.Sprintf("%s: %U, runtime %U", c.Op, got, cps))
				}
				continue
			}
			if err := json.Unmarshal(c.Out, &want); err != nil {
				t.Fatalf("%s %q: %v", c.Op, c.In, err)
			}
		}
		var got string
		var gotErr error
		switch c.Op {
		case "lower":
			got = jsbuiltin.ToLowerCase(c.In)
		case "upper":
			got = jsbuiltin.ToUpperCase(c.In)
		case "number":
			got = numStr(jsbuiltin.ToNumber(c.In))
		case "parseInt":
			got = numStr(jsbuiltin.ParseInt(c.In))
		case "atob":
			var b []byte
			b, gotErr = jsbuiltin.Atob(c.In)
			got = hex.EncodeToString(b)
		case "btoa":
			got, gotErr = jsbuiltin.Btoa(c.In)
		case "encodeURIComponent":
			got, gotErr = jsbuiltin.EncodeURIComponent(c.In)
		case "encodeURI":
			got, gotErr = jsbuiltin.EncodeURI(c.In)
		default:
			t.Fatalf("unknown op %q", c.Op)
		}
		switch {
		case c.Err != "" && gotErr == nil:
			diffs = append(diffs, fmt.Sprintf("%s(%q): %q, runtime throws %s", c.Op, c.In, got, c.Err))
		case c.Err != "" && !errMatches(c.Op, c.Err, gotErr):
			diffs = append(diffs, fmt.Sprintf("%s(%q): error %v, runtime %s", c.Op, c.In, gotErr, c.Err))
		case c.Err == "" && gotErr != nil:
			diffs = append(diffs, fmt.Sprintf("%s(%q): error %v, runtime %q", c.Op, c.In, gotErr, want))
		case c.Err == "" && got != want:
			diffs = append(diffs, fmt.Sprintf("%s(%q): %q, runtime %q", c.Op, c.In, got, want))
		}
	}
	return diffs, n
}

// errMatches holds atob's two error classes to the runtime's: Node words
// them as the package's sentinels do. Another runtime words them its own way
// (workerd one sentence for both, Bun another), so only the class is
// compared there, and only on Node is the message.
func errMatches(op, runtimeErr string, got error) bool {
	if !strings.HasPrefix(runtimeErr, "InvalidCharacterError: ") && op != "encodeURIComponent" && op != "encodeURI" {
		return false
	}
	msg := strings.TrimPrefix(runtimeErr, "InvalidCharacterError: ")
	switch op {
	case "atob":
		if msg == jsbuiltin.ErrInvalidCharacter.Error() || msg == jsbuiltin.ErrNotCorrectlyEncoded.Error() {
			return got.Error() == msg
		}
		return errors.Is(got, jsbuiltin.ErrInvalidCharacter) || errors.Is(got, jsbuiltin.ErrNotCorrectlyEncoded)
	case "btoa":
		return errors.Is(got, jsbuiltin.ErrBtoaRange)
	}
	return got != nil
}

// TestContract passes when exactly the answers listed in
// testdata/known-differences.txt differ: a new difference fails, and so does
// a listed one that stops differing, until its line is removed
// (KNOWN_DIFFERENCES_WRITE=1 rewrites the list; review the diff). Every line
// there is accounted for in the README.
func TestContract(t *testing.T) {
	diffs, n := compareGoldens(t, "testdata")
	if n < 9000 {
		t.Fatalf("thin goldens: %d answers", n)
	}
	const listPath = "testdata/known-differences.txt"
	if os.Getenv("KNOWN_DIFFERENCES_WRITE") != "" {
		if err := os.WriteFile(listPath, []byte(strings.Join(diffs, "\n")+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	raw, err := os.ReadFile(listPath)
	if err != nil {
		t.Fatal(err)
	}
	known := map[string]bool{}
	for _, l := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		if l != "" {
			known[l] = true
		}
	}
	unexpected := 0
	for _, d := range diffs {
		if known[d] {
			delete(known, d)
			continue
		}
		if unexpected++; unexpected <= 40 {
			t.Errorf("new difference: %s", d)
		}
	}
	for l := range known {
		t.Errorf("listed difference no longer differs (remove it): %s", l)
	}
	t.Logf("%d answers compared, %d differ, all listed", n, len(diffs))
}

// TestWatch compares goldens another runtime wrote (WATCH_DIR) and writes
// the differences to WATCH_REPORT. It reports; it does not fail.
func TestWatch(t *testing.T) {
	dir := os.Getenv("WATCH_DIR")
	if dir == "" {
		t.Skip("WATCH_DIR not set (make watch)")
	}
	diffs, n := compareGoldens(t, dir)
	rt, _ := os.ReadFile(filepath.Join(dir, "runtime.json"))
	report := fmt.Sprintf("runtime: %s\n%d answers compared, %d differ\n%s\n",
		strings.Join(strings.Fields(string(rt)), " "), n, len(diffs), strings.Join(diffs, "\n"))
	if out := os.Getenv("WATCH_REPORT"); out != "" {
		if err := os.WriteFile(out, []byte(report), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("%d answers compared, %d differ", n, len(diffs))
}
