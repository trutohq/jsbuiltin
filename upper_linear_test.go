package jsbuiltin_test

import (
	"strings"
	"testing"
	"time"

	"github.com/trutohq/jsbuiltin"
)

// ToUpperCaseOn(RuntimeBun) is one pass however many Bun-unmapped code points
// a string holds. It once recursed once per such code point and copied the
// rest each time: 300 KB of U+0264 took 6 s and 8 MB overflowed the stack.
func TestBunUpperCaseIsLinearInTheSkippedCodePoints(t *testing.T) {
	if testing.Short() {
		t.Skip("timing")
	}
	big := strings.Repeat("ɤ", 1<<19) // 1 MiB
	start := time.Now()
	if got := jsbuiltin.ToUpperCaseOn(jsbuiltin.RuntimeBun, big); got != big {
		t.Fatal("U+0264 was upper-cased; Bun leaves it")
	}
	if d := time.Since(start); d > 100*time.Millisecond {
		t.Errorf("1 MiB of U+0264 took %v, want under 100ms", d)
	}
	// Linear: four times the input costs about four times the time, not
	// sixteen. Measured on mixed input so both paths run.
	unit := strings.Repeat("aɤ", 1<<15)
	timeOf := func(s string) time.Duration {
		best := time.Hour
		for range 3 {
			start := time.Now()
			jsbuiltin.ToUpperCaseOn(jsbuiltin.RuntimeBun, s)
			best = min(best, time.Since(start))
		}
		return best
	}
	small, large := timeOf(unit), timeOf(strings.Repeat(unit, 8))
	if large > 24*small+5*time.Millisecond {
		t.Errorf("8x the input took %v against %v: not linear", large, small)
	}
	if got := jsbuiltin.ToUpperCaseOn(jsbuiltin.RuntimeBun, "aɤßbɤ"); got != "AɤSSBɤ" {
		t.Errorf("mixed = %q", got)
	}
}
