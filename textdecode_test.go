package jsbuiltin_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/trutohq/jsbuiltin"
)

// TestDecodeUTF8MatchesTextDecoder walks testdata/textdecode-golden.json,
// which is the runtime's own `new TextDecoder('utf-8').decode` over each byte
// sequence (testdata/textdecode.ts). The BOM cases are REVIEW-SLICE7 A-44:
// "one leading BOM is removed" fails against string(bytes).
func TestDecodeUTF8MatchesTextDecoder(t *testing.T) {
	raw, err := os.ReadFile("testdata/textdecode-golden.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name    string `json:"name"`
		Bytes   []int  `json:"bytes"`
		Decoded string `json:"decoded"`
	}
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) < 15 {
		t.Fatalf("the golden has %d cases; it had 15 when it was captured", len(cases))
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			b := make([]byte, len(c.Bytes))
			for i, v := range c.Bytes {
				b[i] = byte(v)
			}
			if got := jsbuiltin.DecodeUTF8(b); got != c.Decoded {
				t.Fatalf("DecodeUTF8(% x) = %q, want %q", b, got, c.Decoded)
			}
		})
	}
}
