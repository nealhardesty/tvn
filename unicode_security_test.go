package main

import (
	"testing"

	"golang.org/x/text/unicode/norm"
)

// Regression test for GO-2026-5970 / CVE-2026-56852. Older x/text
// versions could return an empty segment without advancing on malformed
// UTF-8; this input was reported in golang/go#73945.
func TestNormIterMakesProgressOnMalformedUTF8(t *testing.T) {
	const malformed = "\xf0\xd9\x95"
	for _, tc := range []struct {
		name string
		form norm.Form
	}{
		{"NFC", norm.NFC},
		{"NFD", norm.NFD},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := tc.form.String(malformed)
			var iter norm.Iter
			iter.InitString(tc.form, input)
			for !iter.Done() {
				before := iter.Pos()
				iter.Next()
				if iter.Pos() <= before {
					t.Fatalf("normalizer made no progress at byte %d of %x", before, []byte(input))
				}
			}
		})
	}
}

func TestNormalizeUnicodeMalformedUTF8(t *testing.T) {
	// Linux filesystems permit filenames with malformed UTF-8 byte sequences.
	// Normalization must complete without hanging or panicking.
	_ = normalizeUnicode("episode.\xf0\xd9\x95.mkv")
}

func TestNormalizeUnicodeAccents(t *testing.T) {
	const input = "Caf\u00e9.S01E01.mkv"
	const want = "Cafe.S01E01.mkv"
	if got := normalizeUnicode(input); got != want {
		t.Fatalf("normalizeUnicode(%q) = %q, want %q", input, got, want)
	}
}
