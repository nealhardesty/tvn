package main

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/norm"
)

// yearInParens matches a 4-digit year in parentheses, e.g. "(2025)".
var yearInParens = regexp.MustCompile(`\(\d{4}\)`)

// cleanSeriesName converts parsed series name to a clean form.
// Replaces dots, underscores with spaces, strips year-in-parens, trims, and title-cases.
func cleanSeriesName(name string) string {
	name = strings.Map(func(r rune) rune {
		switch r {
		case '.', '_':
			return ' '
		}
		return r
	}, name)
	// Strip year-in-parens like "(2025)" that appear in many download filenames
	name = yearInParens.ReplaceAllString(name, "")
	space := regexp.MustCompile(`\s+`)
	name = space.ReplaceAllString(name, " ")
	name = strings.TrimSpace(name)
	// Title-case each word
	name = cases.Title(language.English).String(name)
	return name
}

// SanitizeFilename applies configured sanitization rules to a filename.
func SanitizeFilename(name string, cfg *Config) string {
	if cfg.WindowsSafeFilenames {
		for _, c := range []string{":", "\"", "<", ">", "|", "?", "*"} {
			name = strings.ReplaceAll(name, c, cfg.ReplaceInvalidCharactersWith)
		}
		// Replace path separators
		name = strings.ReplaceAll(name, "/", cfg.ReplaceInvalidCharactersWith)
		name = strings.ReplaceAll(name, "\\", cfg.ReplaceInvalidCharactersWith)
	}

	if cfg.CustomFilenameCharacterBlacklist != "" {
		for _, c := range cfg.CustomFilenameCharacterBlacklist {
			name = strings.ReplaceAll(name, string(c), cfg.ReplaceInvalidCharactersWith)
		}
	}

	if cfg.NormalizeUnicodeFilenames {
		name = normalizeUnicode(name)
	}

	if cfg.LowercaseFilename {
		name = strings.ToLower(name)
	} else if cfg.TitlecaseFilename {
		name = cases.Title(language.English).String(name)
	}

	return name
}

// normalizeUnicode replaces accented characters with ASCII equivalents.
func normalizeUnicode(s string) string {
	// NFD decomposition separates base characters from combining marks
	result := norm.NFD.String(s)
	// Remove combining marks (category Mn)
	var b strings.Builder
	for _, r := range result {
		if !unicode.Is(unicode.Mn, r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// applyReplacements applies a list of replacements to a string.
func applyReplacements(s string, replacements []Replacement) string {
	for _, r := range replacements {
		if r.IsRegex {
			re, err := regexp.Compile(r.Match)
			if err != nil {
				continue
			}
			s = re.ReplaceAllString(s, r.Replacement)
		} else {
			s = strings.ReplaceAll(s, r.Match, r.Replacement)
		}
	}
	return s
}
