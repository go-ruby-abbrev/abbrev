<p align="center"><img src="https://raw.githubusercontent.com/go-ruby-abbrev/brand/main/social/go-ruby-abbrev-abbrev.png" alt="go-ruby-abbrev/abbrev" width="720"></p>

# abbrev — go-ruby-abbrev

[![Docs](https://img.shields.io/badge/docs-mkdocs--material-DC2626)](https://go-ruby-abbrev.github.io/docs/)
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.26.4%2B-00ADD8)](https://go.dev/dl/)
[![Coverage](https://img.shields.io/badge/coverage-100%25-1a7f37)](#tests--coverage)

**A pure-Go (no cgo) reimplementation of Ruby's [`abbrev`](https://docs.ruby-lang.org/en/master/Abbrev.html)
standard library** — MRI 4.0.5's `Abbrev.abbrev` and the `Array#abbrev` core
extension. Given a set of words it computes the set of *unambiguous
abbreviations*: every prefix that identifies exactly one word, plus each full
word. It is a faithful, byte-for-byte port of upstream `abbrev.rb`
(Akinori MUSHA) — **without any Ruby runtime**.

It is the `abbrev` backend for
[go-embedded-ruby](https://github.com/go-embedded-ruby/ruby), but is a
**standalone, reusable** module with no dependency on the Ruby runtime — a
sibling of [go-ruby-yaml](https://github.com/go-ruby-yaml/yaml) (Psych),
[go-ruby-regexp](https://github.com/go-ruby-regexp/regexp) (the Onigmo engine)
and [go-ruby-erb](https://github.com/go-ruby-erb/erb) (the ERB compiler).

## Install

```sh
go get github.com/go-ruby-abbrev/abbrev
```

## Usage

```go
import "github.com/go-ruby-abbrev/abbrev"

abbrev.Abbrev([]string{"ruby", "rules"})
// => map[string]string{
//      "rub":   "ruby",  "ruby":  "ruby",
//      "rule":  "rules", "rules": "rules",
//    }
// "r" and "ru" are omitted: each is a prefix of both words (ambiguous).

abbrev.Abbrev([]string{"car", "cone"}, "ca")
// => map[string]string{"ca": "car", "car": "car"}
// The optional prefix keeps only words starting with it.
```

This is the idiomatic Go shape of both Ruby entry points:

| Ruby                                | Go                                  |
| ----------------------------------- | ----------------------------------- |
| `Abbrev.abbrev(words)`              | `abbrev.Abbrev(words)`              |
| `Abbrev.abbrev(words, prefix)`      | `abbrev.Abbrev(words, prefix)`      |
| `words.abbrev` (`Array#abbrev`)     | `abbrev.Abbrev(words)`              |
| `words.abbrev(prefix)`              | `abbrev.Abbrev(words, prefix)`      |

## MRI fidelity

The port reproduces MRI 4.0.5 exactly, including its edge cases:

- **Ambiguous prefixes are dropped.** A prefix shared by two or more words never
  appears; each full word always maps to itself.
- **Empty words** contribute no prefixes but still map to themselves:
  `Abbrev([]string{""})` → `{"": ""}`.
- **Duplicate words** collapse: `Abbrev([]string{"dog", "dog"})` → `{"dog": "dog"}`.
- **The prefix is a literal string, not a pattern** (MRI anchors it as
  `/\A<quoted>/`): `Abbrev([]string{"a.b", "axb"}, "a.")` → `{"a.b": "a.b", "a.": "a.b"}`.
- **Multibyte words split on characters**, not bytes, matching Ruby's
  `String#[]` semantics: `Abbrev([]string{"café", "cane"})` includes `"caf"`.

The optional Ruby `pattern` may also be a `Regexp`; this library models the
common String-prefix case (which is what `rbgo` binds), so a host that needs the
regexp form filters the word list before calling `Abbrev`.

## Tests & coverage

The suite is driven to **100% coverage** by deterministic, runtime-free cases,
plus a differential **oracle** that runs the same inputs through the live `ruby`
binary and asserts byte-identical maps. The oracle skips itself where `ruby` is
absent (the Windows lane, the qemu cross-arch lanes) and is gated on
`RUBY_VERSION >= "4.0"`, so the deterministic suite alone keeps the gate green
everywhere. CI validates the library on the six supported 64-bit architectures
(amd64/arm64 natively; riscv64/loong64/ppc64le/s390x under qemu) and on
Linux/macOS/Windows.

```sh
go test ./...
```

## License

BSD-3-Clause — see [LICENSE](LICENSE). Copyright (c) 2026, the
go-ruby-abbrev/abbrev authors.
