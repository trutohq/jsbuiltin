// The goldens: every expected value under testdata/ comes from EXECUTING the
// runtime this file runs on, never from a hand-written table.
//
//   node oracle/generate.mjs [outDir]     (the contract: Node, V8; `make goldens`)
//   bun  oracle/generate.mjs <tmpDir>     (the watch layer runs it on Bun too)
//
// Writes:
//   numbers.jsonl    float64 bits -> Number::toString (and JSON.stringify),
//                    the same cases jsorder's generator writes (same PRNG).
//   builtins.jsonl   {op, in, out} or {op, in, err}: the other operations.
//   runtime.json     which runtime and version answered.

import { writeFileSync, mkdirSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = join(dirname(fileURLToPath(import.meta.url)), "..");
const outDir = process.argv[2] ?? join(root, "testdata");
mkdirSync(outDir, { recursive: true });

const dv = new DataView(new ArrayBuffer(8));
function bitsOf(n) {
  dv.setFloat64(0, n);
  return dv.getBigUint64(0).toString(16).padStart(16, "0");
}

// ---------------------------------------------------------------- numbers
// Copied from jsorder's internal/gen/generate.mjs, so the two files agree.
const adversarial = [
  0, -0, 1, -1, 2, 10, 100, 1000,
  0.1, 0.2, 0.1 + 0.2, 0.3, 1 / 3, 2 / 3, 4.35, 1.005,
  // The exponential-notation thresholds. ECMAScript switches at n > 21 and
  // n <= -6; Go switches somewhere else entirely.
  1e15, 1e16, 1e17, 1e18, 1e19, 1e20, 1e21, 1e22, 1e23,
  -1e20, -1e21,
  1e-1, 1e-5, 1e-6, 1e-7, 1e-8, 1e-10, 1e-20,
  9.999999999999999e20, 1.0000000000000001e21,
  // Denormals and the extremes.
  5e-324, 1e-323, 2.2250738585072014e-308, 2.2250738585072009e-308,
  1.7976931348623157e308, -1.7976931348623157e308,
  // The integer-precision boundary.
  2 ** 53 - 1, 2 ** 53, 2 ** 53 + 2, -(2 ** 53),
  2 ** 31, 2 ** 32 - 1, 2 ** 32, 2 ** 63,
  // Values with long shortest-representations.
  1.2345678901234567e-5, 1.23e-10, 123456789012345678901234,
  5.960464477539063e-8, 1.7e-308,
  // Non-finite: Number::toString and JSON.stringify disagree about these.
  NaN, Infinity, -Infinity,
];

// A deterministic PRNG, so re-running the generator does not churn the goldens.
let seed = 0x2545f491;
function rnd() {
  seed ^= seed << 13; seed >>>= 0;
  seed ^= seed >>> 17;
  seed ^= seed << 5; seed >>>= 0;
  return seed / 0x100000000;
}

const random = [];
// Random doubles from raw bit patterns: hits denormals, huge exponents and
// digit strings no human would think to write.
for (let i = 0; i < 3000; i++) {
  const hi = Math.floor(rnd() * 0x100000000);
  const lo = Math.floor(rnd() * 0x100000000);
  dv.setUint32(0, hi);
  dv.setUint32(4, lo);
  random.push(dv.getFloat64(0));
}
// Random values across the decimal ranges that real payloads actually contain.
for (let i = 0; i < 2000; i++) {
  const exp = Math.floor(rnd() * 45) - 22;
  random.push((rnd() * 2 - 1) * 10 ** exp);
}
// Random integers, the common case.
for (let i = 0; i < 1000; i++) {
  random.push(Math.floor(rnd() * 2 ** Math.floor(rnd() * 60)));
}

const numberLines = [];
for (const n of [...adversarial, ...random]) {
  numberLines.push(JSON.stringify({
    bits: bitsOf(n),
    str: String(n),              // Number::toString
    json: JSON.stringify(n) ?? "null",  // stringify substitutes null for non-finite
  }));
}
writeFileSync(join(outDir, "numbers.jsonl"), numberLines.join("\n") + "\n");

// ---------------------------------------------------------------- builtins

const lines = [];
const add = (op, input, fn) => {
  try {
    lines.push(JSON.stringify({ op, in: input, out: fn(input) }));
  } catch (e) {
    lines.push(JSON.stringify({ op, in: input, err: `${e.name}: ${e.message}` }));
  }
};
const isSurrogate = (c) => c >= 0xd800 && c <= 0xdfff;

// Whitespace: every code point, one line listing those `\s` matches and
// trim() removes (the two must agree; the Go test checks both lists).
{
  const ws = [], trimmed = [];
  for (let c = 0; c <= 0x10ffff; c++) {
    if (isSurrogate(c)) continue;
    const s = String.fromCodePoint(c);
    if (/^\s$/.test(s)) ws.push(c);
    if (("a" + s + "a").trim() === "a" + s + "a" && s.trim() === "") trimmed.push(c);
  }
  lines.push(JSON.stringify({ op: "whitespace", in: "", out: ws }));
  lines.push(JSON.stringify({ op: "trim-whitespace", in: "", out: trimmed }));
}

// Case mapping: every code point whose mapping is not itself, plus strings
// whose answer depends on context (final sigma, dotted I).
for (let c = 0; c <= 0x10ffff; c++) {
  if (isSurrogate(c)) continue;
  const s = String.fromCodePoint(c);
  if (s.toLowerCase() !== s) add("lower", s, (x) => x.toLowerCase());
  if (s.toUpperCase() !== s) add("upper", s, (x) => x.toUpperCase());
}
for (const s of [
  "S14-İΣ.example", "Aİ", "İA", "ABΣ", "X-Σ", "Hello WÖRLD",
  "AΣ B", "AΣ.b", "ZZ꟎", "Σ", "ΣΣ", "AΣ́", "ÁΣ",
  "straße", "ŉ", "ﬀx", "ẞ", "i̇", "ǅ", "", "abc", "ABC",
]) {
  add("lower", s, (x) => x.toLowerCase());
  add("upper", s, (x) => x.toUpperCase());
}

// Number(string) and parseInt(string, 10). The answer is String(result), so
// -0 shows as "0": the Go test compares Object.is through `sign`.
const numberInputs = [
  "", " ", "0", "-0", "+0", "1", "-1", "+1", "01", "00012", "1.5", ".5", "5.", ".", "-.5", "+.5",
  "1e3", "1E3", "1e+3", "1e-3", "1e", "1e+", "e3", "1.2.3", "12abc", "abc", "0x10", "0X1f", "-0x10",
  "0b101", "0B11", "0o17", "0O7", "0b2", "0o8", "0xg", "0x", "Infinity", "-Infinity", "+Infinity",
  "infinity", "INFINITY", "NaN", "1_000", "1,000", " 42 ", "\t42\n", " 42　", "﻿42",
  "\u008542", "᠎42", "42 ", "1" + "0".repeat(400), "-1" + "0".repeat(400),
  "0." + "0".repeat(400) + "1", "9007199254740993", "123456789012345678901234567890",
  "1e1000", "-1e1000", "1e-1000", "0.1", "0.30000000000000004", "99999999999999999999",
  "  -12.5e-2xyz", "١٢", "１", "1 2", "--1", "+-1", "0x1p3", "1n", "1e3.5",
];
const numStr = (n) => (Object.is(n, -0) ? "-0" : String(n));
for (const s of numberInputs) {
  add("number", s, (x) => numStr(Number(x)));
  add("parseInt", s, (x) => numStr(parseInt(x, 10)));
}

// atob: the accepted strings' bytes as hex, or the DOMException.
const hex = (s) => [...s].map((c) => c.charCodeAt(0).toString(16).padStart(2, "0")).join("");
for (const s of [
  "", "YQ==", "YQ", "YQ=", "YQ===", "YWJj", "YWI=", "YWI", "  Y\nW\tJj  ", "Y\fW\rJ j", "a", "aaaaa",
  "!!!!", "Y=Q=", "=", "==", "===", "Y===", "YQ==YQ==", "YR==", "YR", "Zm9vYmFy", "Zm9vYmE=",
  "Zm9vYg==", "Zm9vYg", "Zm9vY", "-_-_", "++//", "éééé", "YQ ==", "YQ\v==",
  "AAAA====", "AAA=A", "Y Q = =", "YQ==\n", "\nYQ==",
]) {
  add("atob", s, (x) => hex(atob(x)));
}
// btoa over a JavaScript string: code units above 0xFF throw.
for (const s of ["", "a", "ab", "abc", "é", "ÿ\u0000", "Ā", "😀", "foobar"]) {
  add("btoa", s, (x) => btoa(x));
}

// encodeURIComponent / encodeURI (a lone surrogate throws a URIError; the
// generator keeps to well-formed strings, which JSON can carry).
for (const s of [
  "", "abc", "a b", "a+b", "a&b=c", "é", "😀", "-_.!~*'()", ";/?:@&=+$,#",
  "[]", "%", "%41", "\u0000", "\u007f", "\u0080", "߿", "ࠀ", "￿", "\u{10ffff}",
]) {
  add("encodeURIComponent", s, (x) => encodeURIComponent(x));
  add("encodeURI", s, (x) => encodeURI(x));
}

writeFileSync(join(outDir, "builtins.jsonl"), lines.join("\n") + "\n");
writeFileSync(
  join(outDir, "runtime.json"),
  JSON.stringify(
    typeof Bun !== "undefined"
      ? { runtime: "bun", version: Bun.version }
      : { runtime: "node", version: process.version, v8: process.versions.v8, icu: process.versions.icu, unicode: process.versions.unicode },
    null, 2,
  ) + "\n",
);
console.error(`numbers=${numberLines.length} builtins=${lines.length}`);
