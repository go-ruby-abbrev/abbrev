// Copyright (c) the go-ruby-abbrev/abbrev authors
//
// SPDX-License-Identifier: BSD-3-Clause

// Package abbrev is a pure-Go (no cgo) reimplementation of Ruby's `abbrev`
// standard library — MRI's [Abbrev.abbrev] and the [Array#abbrev] core
// extension.
//
// Given a set of words it returns the set of unambiguous abbreviations: every
// prefix of a word that is a prefix of exactly one word in the set, mapped to
// that word, plus each full word mapped to itself. Prefixes shared by two or
// more words are ambiguous and omitted.
//
//	abbrev.Abbrev([]string{"ruby", "rules"})
//	// => {"rub":"ruby", "ruby":"ruby", "rule":"rules", "rules":"rules"}
//
// The optional prefix argument restricts the result to words that start with
// the given string (only those words, and only abbreviations at least as long
// as the prefix, appear in the output) — mirroring MRI's String-pattern case.
//
// The algorithm is a faithful port of upstream abbrev.rb (Akinori MUSHA), so it
// reproduces MRI 4.0.5 byte-for-byte, including its handling of empty words,
// duplicate words, and the prefix filter. It is the abbrev backend for
// go-embedded-ruby but is a standalone, runtime-free module.
package abbrev

// Abbrev returns the set of unambiguous abbreviations for words as a map from
// each abbreviation (and each full word) to the full word it identifies.
//
// It is a direct port of MRI's Abbrev.abbrev. An optional prefix may be passed;
// when present, only words starting with prefix are considered and only
// abbreviations that themselves start with prefix appear. Matching MRI, when a
// prefix is given the full word always maps to itself even if the prefix equals
// the whole word; passing more than one prefix uses only the first (extra
// arguments are ignored, as Ruby takes a single pattern).
//
// The returned map is never nil; an empty word set (or one fully filtered out)
// yields an empty map.
func Abbrev(words []string, prefix ...string) map[string]string {
	table := make(map[string]string)
	seen := make(map[string]int)

	var pat string
	var havePat bool
	if len(prefix) > 0 {
		pat = prefix[0]
		havePat = true
	}

	for _, word := range words {
		// MRI: `next if word.empty?` — empty words generate no prefixes here
		// (but may still be added to the table by the final loop below).
		if word == "" {
			continue
		}
		runes := []rune(word)
		// word.size.downto(1) { |len| abbrev = word[0...len] }
		for size := len(runes); size >= 1; size-- {
			ab := string(runes[:size])

			if havePat && !hasPrefix(ab, pat) {
				continue
			}

			seen[ab]++
			switch seen[ab] {
			case 1:
				table[ab] = word
			case 2:
				delete(table, ab)
			default:
				// Seen three or more times already: every shorter prefix of
				// this word has likewise been seen, so stop walking down.
				size = 0 // break out of the downto loop (loop ends after dec)
			}
		}
	}

	for _, word := range words {
		if havePat && !hasPrefix(word, pat) {
			continue
		}
		table[word] = word
	}

	return table
}

// hasPrefix reports whether s begins with prefix. MRI treats a String pattern as
// the anchored regexp /\A<quoted prefix>/, i.e. a literal prefix test, so the
// prefix is never interpreted as a pattern.
func hasPrefix(s, prefix string) bool {
	if len(prefix) > len(s) {
		return false
	}
	return s[:len(prefix)] == prefix
}
