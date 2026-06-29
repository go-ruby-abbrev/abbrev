// Copyright (c) the go-ruby-abbrev/abbrev authors
//
// SPDX-License-Identifier: BSD-3-Clause

package abbrev

import (
	"reflect"
	"testing"
)

// TestAbbrev drives the deterministic, ruby-free behaviour of Abbrev across the
// documented shapes. These cases alone keep coverage at 100%, so the no-ruby CI
// lanes (Windows, the qemu cross-arch lanes) still pass the gate. Each `want` is
// the exact map MRI 4.0.5 returns for the same inputs (see oracle_test.go).
func TestAbbrev(t *testing.T) {
	cases := []struct {
		name   string
		words  []string
		prefix []string
		want   map[string]string
	}{
		{
			name:  "two words drop ambiguous",
			words: []string{"ruby", "rules"},
			want: map[string]string{
				"rub": "ruby", "ruby": "ruby",
				"rul": "rules", "rule": "rules", "rules": "rules",
			},
		},
		{
			name:  "car cone",
			words: []string{"car", "cone"},
			want: map[string]string{
				"ca": "car", "car": "car",
				"co": "cone", "con": "cone", "cone": "cone",
			},
		},
		{
			name:  "single word maps every prefix",
			words: []string{"ruby"},
			want: map[string]string{
				"r": "ruby", "ru": "ruby", "rub": "ruby", "ruby": "ruby",
			},
		},
		{
			name:  "empty list",
			words: []string{},
			want:  map[string]string{},
		},
		{
			name:  "duplicate word",
			words: []string{"dog", "dog"},
			want:  map[string]string{"dog": "dog"},
		},
		{
			name:  "triple duplicate hits the break path",
			words: []string{"dog", "dog", "dog"},
			want:  map[string]string{"dog": "dog"},
		},
		{
			name:  "dup word sharing a prefix with a shorter word",
			words: []string{"do", "dog", "dog"},
			want:  map[string]string{"do": "do", "dog": "dog"},
		},
		{
			name:  "empty word maps to itself in final pass",
			words: []string{""},
			want:  map[string]string{"": ""},
		},
		{
			name:  "empty word alongside a real word",
			words: []string{"", "a"},
			want:  map[string]string{"": "", "a": "a"},
		},
		{
			name:   "prefix filters and keeps only matching words",
			words:  []string{"car", "cone"},
			prefix: []string{"ca"},
			want:   map[string]string{"ca": "car", "car": "car"},
		},
		{
			name:   "prefix excludes non-matching words entirely",
			words:  []string{"foo", "bar"},
			prefix: []string{"f"},
			want:   map[string]string{"f": "foo", "fo": "foo", "foo": "foo"},
		},
		{
			name:   "prefix equal to a whole word still maps full words",
			words:  []string{"abc", "abd"},
			prefix: []string{"ab"},
			want:   map[string]string{"abc": "abc", "abd": "abd"},
		},
		{
			name:   "prefix longer than any word yields empty",
			words:  []string{"ab"},
			prefix: []string{"abc"},
			want:   map[string]string{},
		},
		{
			name:   "prefix is a literal string not a regexp",
			words:  []string{"a.b", "axb"},
			prefix: []string{"a."},
			want:   map[string]string{"a.b": "a.b", "a.": "a.b"},
		},
		{
			name:   "empty prefix matches everything",
			words:  []string{"", "a"},
			prefix: []string{""},
			want:   map[string]string{"": "", "a": "a"},
		},
		{
			name:  "multibyte words split on characters not bytes",
			words: []string{"café", "cane"},
			want: map[string]string{
				"caf": "café", "café": "café",
				"can": "cane", "cane": "cane",
			},
		},
		{
			name:   "explicit prefix arg given to Abbrev",
			words:  []string{"ruby", "rules"},
			prefix: []string{"rub"},
			want:   map[string]string{"rub": "ruby", "ruby": "ruby"},
		},
		{
			name:   "extra prefix args after the first are ignored",
			words:  []string{"car", "cone"},
			prefix: []string{"ca", "ignored", "also-ignored"},
			want:   map[string]string{"ca": "car", "car": "car"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Abbrev(c.words, c.prefix...)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("Abbrev(%q, %q) = %v, want %v", c.words, c.prefix, got, c.want)
			}
		})
	}
}

// TestAbbrevReturnsNonNil documents that the result is always a usable (non-nil)
// map even for an empty input, so callers can range/index it unconditionally.
func TestAbbrevReturnsNonNil(t *testing.T) {
	if Abbrev(nil) == nil {
		t.Fatal("Abbrev(nil) returned a nil map")
	}
}

// TestHasPrefix exercises hasPrefix directly, including the over-long-prefix
// branch that the table cases reach only indirectly.
func TestHasPrefix(t *testing.T) {
	cases := []struct {
		s, prefix string
		want      bool
	}{
		{"car", "ca", true},
		{"car", "car", true},
		{"car", "card", false}, // prefix longer than s
		{"car", "x", false},
		{"", "", true},
		{"a", "", true},
	}
	for _, c := range cases {
		if got := hasPrefix(c.s, c.prefix); got != c.want {
			t.Errorf("hasPrefix(%q, %q) = %v, want %v", c.s, c.prefix, got, c.want)
		}
	}
}
