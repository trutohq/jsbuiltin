# Handoff

What each consumer changes to use this module and delete its copy. Written
2026-10-31; gotruto and goenvoy were not edited (the brief).

The module is not published; consumers use a replace:

```
require github.com/trutohq/jsbuiltin v0.0.0
replace github.com/trutohq/jsbuiltin => ../jsbuiltin   // adjust the depth
```

## jsorder (done)

`number.go` moved here; jsorder's `AppendNumber` / `FormatNumber` call
`jsbuiltin.AppendNumber` / `jsbuiltin.FormatNumber`, so its API is unchanged.
jsorder consumers add the replace above, because a replace in jsorder's go.mod
does not reach them.

## truto-jsonata-go (done)

`internal/jsonx` deleted; `jsstring.go`, `jsnum.go`, `fn_crypto.go` call this
module.

## gotruto

| there | change |
|---|---|
| `internal/jsonx` `lowercase.go`, `utf16.go`, `textdecode.go`, `uri.go` | delete; callers use `jsbuiltin.ToLowerCase`, `UTF16Len`, `UTF16Units`, `UTF16Prefix`, `UTF16Unit`, `UTF16String`, `HasLoneSurrogate`, `CompareUTF16`, `DecodeUTF8`, `WellFormedUTF8`, `EncodeURIComponent`, `EncodeURI` (same names and signatures) |
| `internal/jsonx` `Trim`, `IsJSSpace` | `jsbuiltin.Trim`, `jsbuiltin.IsWhiteSpace` (same set) |
| `internal/jsonx` `NumberValue` | `jsbuiltin.NumberValue` |
| `internal/jsonx` `IsFalsy`, `StringOf`, `Omit`, `Parse`, `Numbers` | stay: they are glue over jsorder's value model, and jsbuiltin cannot import jsorder (jsorder imports it) |
| `internal/qsx` `ToNumber`, `ParseInt`, `IsJSWhitespace` | `jsbuiltin.ToNumber`, `jsbuiltin.ParseInt`, `jsbuiltin.IsWhiteSpace` (qs-go already uses them; see its HANDOFF) |
| `internal/qsx` `parseint_differential_test.go` | its all-code-points whitespace check is `TestContract`'s whitespace line here; keep gotruto's only if it wants workerd measured as well |
| `internal/cursor` `atob` / `Atob` | `jsbuiltin.Atob`. Same accepted strings and bytes; map either error to `ErrBase64` / `WorkerdAtobMessage` as today |
| `internal/analytics` `atob` | `jsbuiltin.Atob` (`ok` is `err == nil`) |
| `internal/saml` `atob` | `jsbuiltin.Atob`, keeping its `atobError` wording |
| `internal/cursor` `btoaBytes` (`ginger.go`) | `jsbuiltin.Btoa` (`ErrBtoaRange` for a code unit above 0xFF) |
| `go.mod` | require and replace this module |

## goenvoy

Named in gotruto's notes, not read here: `internal/secrets/crypto.go`
`jsAtob`, `internal/ginger/pagination.go` (a stricter base64, probably wrong),
`internal/expr.IsJSWhitespace`, `internal/connector/httpx/httpx.go` and
`internal/scheduler/cron_pattern.go` parseInt copies. Each becomes the call
here; the owner of goenvoy checks each against this module's goldens first.
