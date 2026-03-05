package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// EpisodeType indicates the type of episode parsed.
type EpisodeType int

const (
	EpisodeStandard  EpisodeType = iota
	EpisodeDateBased
	EpisodeAnime
)

// ParseResult holds the structured data parsed from a filename.
type ParseResult struct {
	SeriesName     string
	SeasonNumber   int
	HasSeason      bool
	EpisodeNumbers []int

	// Date-based
	Year  int
	Month int
	Day   int

	// Anime
	Group string
	CRC   string

	Extension   string
	BaseName    string // original filename without extension
	EpisodeType EpisodeType
}

// DateString returns the formatted date for date-based episodes.
func (p *ParseResult) DateString() string {
	return fmt.Sprintf("%04d-%02d-%02d", p.Year, p.Month, p.Day)
}

// Parser parses TV episode filenames using configured patterns.
type Parser struct {
	patterns []compiledPattern
	extRe    *regexp.Regexp
}

// NewParser creates a parser from the given config.
func NewParser(cfg *Config) (*Parser, error) {
	var defs []patternDef
	if len(cfg.FilenamePatterns) > 0 {
		// Custom patterns from config
		for _, p := range cfg.FilenamePatterns {
			defs = append(defs, patternDef{
				Regex:       p,
				IsDateBased: hasNamedGroup(p, "year"),
				IsAnime:     hasNamedGroup(p, "group"),
			})
		}
	} else {
		defs = defaultPatternDefs
	}

	patterns, err := compilePatterns(defs)
	if err != nil {
		return nil, fmt.Errorf("compiling patterns: %w", err)
	}

	extRe, err := regexp.Compile(cfg.ExtensionPattern)
	if err != nil {
		return nil, fmt.Errorf("compiling extension pattern: %w", err)
	}

	return &Parser{patterns: patterns, extRe: extRe}, nil
}

// Parse parses a filename and returns structured episode data.
func (p *Parser) Parse(filename string) (*ParseResult, error) {
	base, ext := p.splitExtension(filename)

	for _, pat := range p.patterns {
		indices := pat.re.FindStringSubmatchIndex(base)
		if indices == nil {
			continue
		}

		result := &ParseResult{
			BaseName:  base,
			Extension: ext,
		}

		groups := extractNamedGroups(pat.re, base, indices)

		// Series name
		if sn, ok := groups["seriesname"]; ok {
			result.SeriesName = cleanSeriesName(sn)
		}
		if result.SeriesName == "" {
			continue // skip patterns that don't capture a series name
		}

		if pat.isDateBased {
			result.EpisodeType = EpisodeDateBased
			result.Year, _ = strconv.Atoi(groups["year"])
			result.Month, _ = strconv.Atoi(groups["month"])
			result.Day, _ = strconv.Atoi(groups["day"])
			return result, nil
		}

		if pat.isAnime {
			result.EpisodeType = EpisodeAnime
			result.Group = groups["group"]
			result.CRC = groups["crc"]
		}

		// Season number
		if sn, ok := groups["seasonnumber"]; ok && sn != "" {
			result.SeasonNumber, _ = strconv.Atoi(sn)
			result.HasSeason = true
		}

		// Episode numbers
		if epStart, ok := groups["episodenumberstart"]; ok {
			start, _ := strconv.Atoi(epStart)
			end, _ := strconv.Atoi(groups["episodenumberend"])
			for i := start; i <= end; i++ {
				result.EpisodeNumbers = append(result.EpisodeNumbers, i)
			}
		} else if ep, ok := groups["episodenumber"]; ok {
			firstEp, _ := strconv.Atoi(ep)

			if pat.multiEpSuffix != nil {
				result.EpisodeNumbers = p.extractMultiEpisodes(base, pat, indices, firstEp)
			} else {
				result.EpisodeNumbers = []int{firstEp}
			}
		}

		// Check for numbered episode groups: episodenumber1, episodenumber2, ...
		for i := 1; i <= 10; i++ {
			key := fmt.Sprintf("episodenumber%d", i)
			if v, ok := groups[key]; ok && v != "" {
				ep, _ := strconv.Atoi(v)
				result.EpisodeNumbers = append(result.EpisodeNumbers, ep)
			}
		}

		if len(result.EpisodeNumbers) == 0 && result.EpisodeType != EpisodeDateBased {
			continue // no episode number found, try next pattern
		}

		return result, nil
	}

	return nil, fmt.Errorf("no pattern matched: %s", filename)
}

// splitExtension splits a filename into base name and extension.
func (p *Parser) splitExtension(filename string) (string, string) {
	loc := p.extRe.FindStringIndex(filename)
	if loc == nil {
		return filename, ""
	}
	return filename[:loc[0]], filename[loc[0]:]
}

// extractMultiEpisodes finds all episode numbers using the suffix pattern.
func (p *Parser) extractMultiEpisodes(basename string, pat compiledPattern, indices []int, firstEp int) []int {
	// Find position of the episodenumber group
	names := pat.re.SubexpNames()
	epGroupIdx := -1
	for i, name := range names {
		if name == "episodenumber" {
			epGroupIdx = i
			break
		}
	}
	if epGroupIdx < 0 || indices[epGroupIdx*2] < 0 {
		return []int{firstEp}
	}

	epEnd := indices[epGroupIdx*2+1]
	rest := basename[epEnd:]

	episodes := []int{firstEp}
	matches := pat.multiEpSuffix.FindAllStringSubmatch(rest, -1)
	for _, m := range matches {
		if len(m) > 1 {
			ep, _ := strconv.Atoi(m[1])
			episodes = append(episodes, ep)
		}
	}
	return episodes
}

// extractNamedGroups returns a map of named group name -> matched value.
func extractNamedGroups(re *regexp.Regexp, s string, indices []int) map[string]string {
	groups := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if name == "" || i*2 >= len(indices) {
			continue
		}
		start := indices[i*2]
		end := indices[i*2+1]
		if start >= 0 && end >= 0 {
			groups[name] = s[start:end]
		}
	}
	return groups
}

// hasNamedGroup checks if a regex pattern string contains a named group.
func hasNamedGroup(pattern, name string) bool {
	return strings.Contains(pattern, "(?P<"+name+">")
}
