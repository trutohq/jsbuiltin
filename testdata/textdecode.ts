// Captures `new TextDecoder('utf-8').decode(bytes)` — the call behind
// src/common/arrayBufferToString.ts and the keyset cursor
// (src/common/AbstractService.ts:236-238) — for the byte sequences jsonx.DecodeUTF8
// has to agree on: the BOM rule and the maximal-subpart replacement rule.
//
//   cd gotruto/internal/jsonx/testdata && bun textdecode.ts > textdecode-golden.json
const cases: [string, number[]][] = [
  ['empty', []],
  ['ascii', [0x61, 0x62]],
  ['one leading BOM is removed', [0xef, 0xbb, 0xbf, 0x61]],
  ['a BOM alone decodes to the empty string', [0xef, 0xbb, 0xbf]],
  [
    'only ONE leading BOM is removed',
    [0xef, 0xbb, 0xbf, 0xef, 0xbb, 0xbf, 0x61],
  ],
  ['a BOM that is not first stays', [0x61, 0xef, 0xbb, 0xbf]],
  ['a truncated BOM is one replacement', [0xef, 0xbb]],
  ['a truncated BOM then text', [0xef, 0xbb, 0x61]],
  ['a BOM before a JSON object', [0xef, 0xbb, 0xbf, 0x7b, 0x7d]],
  ['a lone continuation byte', [0x80]],
  ['a truncated four-byte sequence is ONE replacement', [0xf0, 0x9f, 0x98]],
  ['an overlong two-byte form is two replacements', [0xc0, 0xaf]],
  ['a surrogate encoded in UTF-8 is three replacements', [0xed, 0xa0, 0x80]],
  ['a BOM then an invalid byte', [0xef, 0xbb, 0xbf, 0xff, 0x61]],
  ['astral', [0xf0, 0x9f, 0x98, 0x80]],
]
const d = new TextDecoder('utf-8')
console.log(
  JSON.stringify(
    cases.map(([name, bytes]) => ({
      name,
      bytes,
      decoded: d.decode(new Uint8Array(bytes)),
    })),
    null,
    2
  )
)
