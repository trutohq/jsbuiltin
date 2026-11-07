// The Bun contract: what Bun answers where the package has a RuntimeBun
// profile, from EXECUTING Bun, never from a hand-written table.
//
//   bun oracle/bun.mjs [outDir]     (default testdata/bun; `make bun-goldens`)
//
// Writes:
//   bun.jsonl      {op, in, out} or {op, err} lines:
//                  - "lower": every code point whose toLowerCase() is not
//                    itself, and the context strings. The list is COMPLETE:
//                    the Go test holds every other code point to itself.
//                  - "uri-error": the message encodeURIComponent and
//                    encodeURI throw for a lone surrogate (JSON cannot carry
//                    the input, so the input is fixed: "\ud800").
//                  - "encodeURIComponent" / "encodeURI": well-formed strings.
//   runtime.json   which runtime and version answered.

import { writeFileSync, mkdirSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

if (typeof Bun === "undefined") {
  console.error("run this with bun: it records Bun's answers");
  process.exit(1);
}
const root = join(dirname(fileURLToPath(import.meta.url)), "..");
const outDir = process.argv[2] ?? join(root, "testdata", "bun");
mkdirSync(outDir, { recursive: true });

const lines = [];
const add = (op, s, f) => lines.push(JSON.stringify({ op, in: s, out: f(s) }));

for (let c = 0; c <= 0x10ffff; c++) {
  if (c >= 0xd800 && c <= 0xdfff) continue;
  const s = String.fromCodePoint(c);
  if (s.toLowerCase() !== s) add("lower", s, (x) => x.toLowerCase());
}
for (const s of [
  "S14-İΣ.example", "Aİ", "İA", "ABΣ", "X-Σ", "Hello WÖRLD",
  "AΣ B", "AΣ.b", "ZZ꟎", "Σ", "ΣΣ", "AΣ́", "ÁΣ",
  "straße", "ŉ", "ﬀx", "ẞ", "i̇", "ǅ", "", "abc", "ABC",
]) {
  add("lower-context", s, (x) => x.toLowerCase());
}

for (const f of [encodeURIComponent, encodeURI]) {
  try {
    f("\ud800");
    lines.push(JSON.stringify({ op: "uri-error", in: f.name, err: "" }));
  } catch (e) {
    lines.push(JSON.stringify({ op: "uri-error", in: f.name, err: `${e.name}: ${e.message}` }));
  }
}
for (const s of [
  "", "abc", "a b", "a+b", "a&b=c", "é", "😀", "-_.!~*'()", ";/?:@&=+$,#",
  "[]", "%", "%41", "\u0000", "\u007f", "\u0080", "߿", "ࠀ", "￿", "\u{10ffff}",
]) {
  add("encodeURIComponent", s, (x) => encodeURIComponent(x));
  add("encodeURI", s, (x) => encodeURI(x));
}

writeFileSync(join(outDir, "bun.jsonl"), lines.join("\n") + "\n");
writeFileSync(join(outDir, "runtime.json"), JSON.stringify({ runtime: "bun", version: Bun.version }, null, 2) + "\n");
console.error(`bun ${Bun.version}: ${lines.length} answers`);
