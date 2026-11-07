package jsbuiltin

// Runtime names the JavaScript runtime whose answer an operation gives, where
// runtimes differ. The package's plain functions are V8's, as Node and workerd
// run it (their contract is testdata/, oracle/generate.mjs on Node). The ...On
// functions take a Runtime: [ToLowerCaseOn], [ToUpperCaseOn], [EncodeURIComponentOn] and
// [EncodeURIOn].
type Runtime int

const (
	// RuntimeV8 is V8, as Node and workerd run it: the plain functions.
	RuntimeV8 Runtime = iota
	// RuntimeBun is Bun 1.3.14 (JavaScriptCore), which runs Envoy. Its
	// contract is testdata/bun/, written by oracle/bun.mjs on Bun.
	RuntimeBun
)
