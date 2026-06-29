// Copyright (c) the go-ruby-abbrev/abbrev authors
//
// SPDX-License-Identifier: BSD-3-Clause

package abbrev

import (
	"encoding/json"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

// rubyBin locates a usable `ruby` once and reports its RUBY_VERSION. The oracle
// tests skip themselves when ruby is absent (the qemu cross-arch lanes and the
// Windows lane) or too old, so the deterministic suite alone drives the 100%
// gate there. The abbrev grammar (Akinori MUSHA's abbrev.rb) is stable, but we
// gate on RUBY_VERSION >= "4.0" so the oracle pins to the supported MRI.
func rubyBin(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("ruby")
	if err != nil {
		t.Skip("ruby not on PATH; skipping MRI oracle")
	}
	out, err := exec.Command(path, "-e", "print RUBY_VERSION").Output()
	if err != nil {
		t.Skipf("could not query RUBY_VERSION: %v", err)
	}
	if ver := string(out); ver < "4.0" {
		t.Skipf("ruby %s < 4.0; skipping MRI oracle", ver)
	}
	return path
}

// rubyAbbrev runs MRI's Abbrev.abbrev for the given words (and optional prefix)
// and returns the resulting hash as a Go map. The script $stdout.binmode's and
// reads from a binmode'd $stdin so Windows text-mode never pollutes the bytes
// (the go-ruby-erb lesson); words and prefix travel as JSON on stdin so no
// shell quoting can corrupt them.
func rubyAbbrev(t *testing.T, bin string, words []string, prefix *string) map[string]string {
	t.Helper()

	in := map[string]any{"words": words}
	if prefix != nil {
		in["prefix"] = *prefix
	}
	payload, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal oracle input: %v", err)
	}

	const script = `
$stdout.binmode
$stdin.binmode
require 'abbrev'
require 'json'
arg = JSON.parse($stdin.read)
words = arg['words']
result = arg.key?('prefix') ? Abbrev.abbrev(words, arg['prefix']) : Abbrev.abbrev(words)
$stdout.write(JSON.generate(result))
`
	cmd := exec.Command(bin, "-e", script)
	cmd.Stdin = strings.NewReader(string(payload))
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ruby error: %v\noutput:\n%s", err, out)
	}

	var got map[string]string
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("decode ruby output %q: %v", out, err)
	}
	return got
}

// TestOracleAgainstMRI runs the same inputs through this package and the live
// ruby binary and asserts byte-identical maps, proving MRI fidelity on every
// platform that ships ruby (the ubuntu/macos CI lanes).
func TestOracleAgainstMRI(t *testing.T) {
	bin := rubyBin(t)

	str := func(s string) *string { return &s }
	cases := []struct {
		name   string
		words  []string
		prefix *string
	}{
		{name: "ruby rules", words: []string{"ruby", "rules"}},
		{name: "car cone", words: []string{"car", "cone"}},
		{name: "single", words: []string{"ruby"}},
		{name: "empty list", words: []string{}},
		{name: "duplicate", words: []string{"dog", "dog"}},
		{name: "triple duplicate", words: []string{"dog", "dog", "dog"}},
		{name: "do dog dog", words: []string{"do", "dog", "dog"}},
		{name: "empty word", words: []string{""}},
		{name: "empty and a", words: []string{"", "a"}},
		{name: "many overlapping", words: []string{"car", "card", "care", "cat", "dog"}},
		{name: "multibyte", words: []string{"café", "cane", "naïve"}},
		{name: "prefix ca", words: []string{"car", "cone"}, prefix: str("ca")},
		{name: "prefix f", words: []string{"foo", "bar"}, prefix: str("f")},
		{name: "prefix equals word", words: []string{"abc", "abd"}, prefix: str("ab")},
		{name: "prefix too long", words: []string{"ab"}, prefix: str("abc")},
		{name: "prefix literal dot", words: []string{"a.b", "axb"}, prefix: str("a.")},
		{name: "prefix empty", words: []string{"", "a"}, prefix: str("")},
		{name: "prefix rub", words: []string{"ruby", "rules"}, prefix: str("rub")},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var got map[string]string
			if c.prefix != nil {
				got = Abbrev(c.words, *c.prefix)
			} else {
				got = Abbrev(c.words)
			}
			want := rubyAbbrev(t, bin, c.words, c.prefix)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Abbrev(%q, prefix=%v) = %v\nMRI = %v", c.words, c.prefix, got, want)
			}
		})
	}
}
