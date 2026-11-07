# jsbuiltin

JavaScript's built-in operations for Go, as V8 and JavaScriptCore compute
them: `Number::toString`, `Number(string)`, `parseInt`, `String.prototype`
`toLowerCase` / `toUpperCase` / `trim` and the whitespace set they share,
UTF-16 length, slicing and ordering, `TextDecoder('utf-8')`, `atob` / `btoa`,
`encodeURIComponent` / `encodeURI`, and the number `JSON.parse` makes of a
JSON number.

```go
import "github.com/trutohq/jsbuiltin"

jsbuiltin.FormatNumber(1e21)          // "1e+21"; 1e-7 is "1e-7", -0 is "0"
n, ok := jsbuiltin.ToNumber(" 0x1f ") // 31, true; "12abc" is NaN (ok false)
b, err := jsbuiltin.Atob("YQ")        // "a": forgiving-base64
```

The JSON value model (`[[OwnPropertyKeys]]` order, `JSON.stringify` bytes)
is [jsorder](../jsorder), which imports its number formatting from here. This
module imports nothing of the org's, so jsorder can.

## Why it exists

The owner's decision gotruto D-0868 (2026-10-30): each JavaScript runtime
piece the Go ports need lives once, in a small module, held to the runtime by
a harness. Before it, these operations had copies in gotruto, goenvoy and
truto-jsonata-go that nothing compared: forgiving-base64 had four in gotruto
and the module alone, the whitespace set six across the org.

## Where the code came from

Moved, with its tests, on 2026-10-31. History stays in the source repositories.

| here | from |
|---|---|
| `number.go` | jsorder `number.go` (b350e12); jsorder keeps `AppendNumber` / `FormatNumber` as wrappers |
| `lowercase.go`, `utf16.go`, `textdecode.go`, `uri.go` and their tests | truto-jsonata-go `internal/jsonx` (0492520), a copy of gotruto `internal/jsonx` at truto e5b134004 |
| `whitespace.go` (`IsWhiteSpace`, `Trim`, `ParseInt`) | gotruto `internal/qsx` `codec.go` `isJSWhitespace` / `parseIntJS` and `internal/jsonx` `IsJSSpace` / `Trim`: one set, two copies, now one |
| `tonumber.go` | gotruto `internal/qsx` `tonumber.go` (via the module's `internal/qs`) |
| `upper.go` | truto-jsonata-go `jsstring.go` `jsToUpperCase` |
| `base64.go` | truto-jsonata-go `fn_crypto.go` `atob` (the one that tells Node's two errors apart); replaces gotruto's `cursor.atob`, `analytics.atob`, `saml.atob`, which accept and refuse the same strings |
| `numbervalue.go` | gotruto `internal/jsonx` `NumberValue` |

## Runtimes

The plain functions give V8's answers, as Node and workerd run it. Where Bun
(JavaScriptCore, which runs Envoy) answers differently, an `...On` function
takes a `Runtime`: `ToLowerCaseOn(RuntimeBun, s)` leaves the 27 code points
Unicode 16 gave a lowercase form as Bun 1.3.14 does, and
`EncodeURIComponentOn` / `EncodeURIOn` with `RuntimeBun` throw Bun's message,
"String contained an illegal UTF-16 sequence.", where V8 says "URI
malformed". `TestBunContract` holds them to `testdata/bun/`, which
`oracle/bun.mjs` writes on Bun (`make bun-goldens`); its lowercase list is
complete, so every code point it does not name must lower to itself.

## How it is checked

**Contract** (`make test`): `oracle/generate.mjs` executes the runtime and
writes every expected value under `testdata/` (`make goldens`;
`testdata/runtime.json` names the runtime: Node 24.18.1, V8 13.6, ICU 78.3,
Unicode 17). `TestContract` compares 9,399 answers: 6,058 `Number::toString`
(the same cases jsorder's generator writes), the whitespace set over every code
point (as `\s` and as `trim`), `toLowerCase` / `toUpperCase` of every code
point that maps and of context-sensitive strings, and `Number`, `parseInt`,
`atob`, `btoa`, `encodeURIComponent`, `encodeURI` over edge cases.
`TextDecoder` has its own golden (`testdata/textdecode.ts`).

The suite passes when exactly the lines in `testdata/known-differences.txt`
differ; a new one fails it, and so does a listed one that stops differing.
Today that is **84**, all one cause (OPEN):

- **Case mapping of characters added in Unicode 16 and 17** (29 `toLowerCase`,
  55 `toUpperCase`; e.g. U+A7CE, U+10D70..U+10D85, U+16EA0..U+16EB8). Node's
  ICU 78 maps them; `toUpperCase` here is x/text's `cases` (Unicode 15.0.0) and
  `toLowerCase`'s general path Go's `unicode.ToLower`, which leaves them too.
  Bun 1.3.14 (Envoy's runtime) does not map them either and agrees with this
  package on all 9,289 of its answers. workerd's ICU (Truto's runtime) was not
  measured. Fixing it is a change of behaviour, so it is not part of the move.

**Watch** (`make watch`, in `make all`): the generator runs again on Bun and on
the installed Node, and `TestWatch` reports what differs into `WATCH.txt`
without failing. 2026-10-31: Node 24.18.1, the 84 above; Bun 1.3.14, none.

Everything else (`make all` is fmt-check, vet, the race suite and the watch):

```sh
export PATH="$HOME/.local/go/bin:$PATH"
make all
```

## Consumers

See HANDOFF.md.
