package main

import (
	"fmt"
	"regexp"
)

// patternDef defines a filename matching pattern.
type patternDef struct {
	Regex         string
	IsDateBased   bool
	IsAnime       bool
	MultiEpSuffix string // regex to find additional episode numbers after first
}

// compiledPattern is a ready-to-use compiled pattern.
type compiledPattern struct {
	re            *regexp.Regexp
	isDateBased   bool
	isAnime       bool
	multiEpSuffix *regexp.Regexp
}

// defaultPatternDefs defines all built-in filename patterns.
// Order matters: first match wins. Most specific patterns first.
var defaultPatternDefs = []patternDef{
	// Date-based: show.2010.01.02.stuff
	{
		Regex:       `(?i)^(?P<seriesname>.+?)[\s._-]+(?P<year>\d{4})[\s._-](?P<month>\d{1,2})[\s._-](?P<day>\d{1,2})(?:[\s._-]|$)`,
		IsDateBased: true,
	},

	// Anime: [Group] Show - 01 - Episode Name [CRC]
	{
		Regex:   `(?i)^\[(?P<group>[^\]]+?)\][\s._-]*(?P<seriesname>.+?)[\s._-]+(?P<episodenumber>\d+)[\s._-]*-[\s._-]*.+[\s._-]*\[(?P<crc>[0-9a-fA-F]{8})\]`,
		IsAnime: true,
	},

	// Anime: [Group] Show - 01 [CRC]
	{
		Regex:   `(?i)^\[(?P<group>[^\]]+?)\][\s._-]*(?P<seriesname>.+?)[\s._-]+(?P<episodenumber>\d+)[\s._-]*\[(?P<crc>[0-9a-fA-F]{8})\]`,
		IsAnime: true,
	},

	// Anime: [Group] Show - 01
	{
		Regex:   `(?i)^\[(?P<group>[^\]]+?)\][\s._-]*(?P<seriesname>.+?)[\s._-]+(?P<episodenumber>\d+)\s*$`,
		IsAnime: true,
	},

	// SxxExx range: show.s01e02-e04 or show.s01e02-04
	{
		Regex: `(?i)^(?P<seriesname>.+?)[\s._-]+[Ss](?P<seasonnumber>\d+)[\s._-]*[Ee](?P<episodenumberstart>\d+)[\s._-]*-[\s._-]*[Ee]?(?P<episodenumberend>\d+)`,
	},

	// SxxExx multi: show.s01e02e03 (detect additional with suffix)
	{
		Regex:         `(?i)^(?P<seriesname>.+?)[\s._-]+[Ss](?P<seasonnumber>\d+)[\s._-]*[Ee](?P<episodenumber>\d+)(?:[\s._-]*[Ee]\d+)+`,
		MultiEpSuffix: `(?i)[Ee](\d+)`,
	},

	// SxxExx single: show.s01e02
	{
		Regex: `(?i)^(?P<seriesname>.+?)[\s._-]+[Ss](?P<seasonnumber>\d+)[\s._-]*[Ee](?P<episodenumber>\d+)`,
	},

	// NxNN range: show.1x02-04
	{
		Regex: `(?i)^(?P<seriesname>.+?)[\s._-]+(?P<seasonnumber>\d+)[xX](?P<episodenumberstart>\d+)-(?P<episodenumberend>\d+)`,
	},

	// NxNN multi: show.1x02x03
	{
		Regex:         `(?i)^(?P<seriesname>.+?)[\s._-]+(?P<seasonnumber>\d+)[xX](?P<episodenumber>\d+)(?:[xX]\d+)+`,
		MultiEpSuffix: `[xX](\d+)`,
	},

	// NxNN single: show.1x02
	{
		Regex: `(?i)^(?P<seriesname>.+?)[\s._-]+(?P<seasonnumber>\d+)[xX](?P<episodenumber>\d+)`,
	},

	// Bracketed: show.[1x09-11]
	{
		Regex: `(?i)^(?P<seriesname>.+?)[\s._-]+\[(?P<seasonnumber>\d+)[xX](?P<episodenumberstart>\d+)-(?P<episodenumberend>\d+)\]`,
	},

	// Bracketed: show.[1x09]
	{
		Regex: `(?i)^(?P<seriesname>.+?)[\s._-]+\[(?P<seasonnumber>\d+)[xX](?P<episodenumber>\d+)\]`,
	},

	// Bracketed: show - [01.09] (season.episode)
	{
		Regex: `(?i)^(?P<seriesname>.+?)[\s._-]+\[(?P<seasonnumber>\d+)\.(?P<episodenumber>\d+)\]`,
	},

	// Bracketed episode only: show - [012]
	{
		Regex: `(?i)^(?P<seriesname>.+?)[\s._-]+\[(?P<episodenumber>\d{2,})\]`,
	},

	// Season Episode spelled out: show Season 01 Episode 20
	{
		Regex: `(?i)^(?P<seriesname>.+?)[\s._-]+season[\s._-]*(?P<seasonnumber>\d+)[\s._-]*episode[\s._-]*(?P<episodenumber>\d+)`,
	},

	// Show - Episode 9999 [S 12 - Ep 131]
	{
		Regex: `(?i)^(?P<seriesname>.+?)[\s._-]+episode[\s._-]*\d+[\s._-]*\[[\s._-]*[Ss][\s._-]*(?P<seasonnumber>\d+)[\s._-]*-?[\s._-]*[Ee]p?[\s._-]*(?P<episodenumber>\d+)\]`,
	},

	// Alternate: Foo - S2 E 02
	{
		Regex: `(?i)^(?P<seriesname>.+?)[\s._-]+[Ss][\s._-]*(?P<seasonnumber>\d+)[\s._-]+[Ee][\s._-]*(?P<episodenumber>\d+)`,
	},

	// Compact SxxExx: show.s0102 (s01e02)
	{
		Regex: `(?i)^(?P<seriesname>.+?)[\s._-]+[Ss](?P<seasonnumber>\d{2})(?P<episodenumber>\d{2})(?:\D|$)`,
	},

	// Part notation: show Part 1 and Part 2
	{
		Regex: `(?i)^(?P<seriesname>.+?)[\s._-]+part[\s._-]*(?P<episodenumberstart>\d+)[\s._-]+and[\s._-]+part[\s._-]*(?P<episodenumberend>\d+)`,
	},

	// Part notation: show Part1
	{
		Regex: `(?i)^(?P<seriesname>.+?)[\s._-]+part[\s._-]*(?P<episodenumber>\d+)`,
	},

	// N of M: show 2 of 6
	{
		Regex: `(?i)^(?P<seriesname>.+?)[\s._-]+(?P<episodenumber>\d+)[\s._-]*of[\s._-]*\d+`,
	},

	// Simple E format: show.e123
	{
		Regex: `(?i)^(?P<seriesname>.+?)[\s._-]+[Ee](?P<episodenumber>\d+)`,
	},
}

// compilePatterns compiles pattern definitions into ready-to-use patterns.
func compilePatterns(defs []patternDef) ([]compiledPattern, error) {
	patterns := make([]compiledPattern, 0, len(defs))
	for i, d := range defs {
		re, err := regexp.Compile(d.Regex)
		if err != nil {
			return nil, fmt.Errorf("pattern %d: %w", i, err)
		}
		var suffix *regexp.Regexp
		if d.MultiEpSuffix != "" {
			suffix, err = regexp.Compile(d.MultiEpSuffix)
			if err != nil {
				return nil, fmt.Errorf("pattern %d suffix: %w", i, err)
			}
		}
		patterns = append(patterns, compiledPattern{
			re:            re,
			isDateBased:   d.IsDateBased,
			isAnime:       d.IsAnime,
			multiEpSuffix: suffix,
		})
	}
	return patterns, nil
}
