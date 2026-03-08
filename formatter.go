package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"
)

// pyFmtRe matches Python-style %(key)s / %(key)02d format placeholders.
var pyFmtRe = regexp.MustCompile(`%\((\w+)\)(?:[0-9]*[a-zA-Z])?`)

// pyKeyField maps tvnamer Python config keys to Go EpisodeData field names.
var pyKeyField = map[string]string{
	"seriesname":   "SeriesName",
	"seasonnumber": "SeasonNumber",
	"episode":      "Episode",
	"episodename":  "EpisodeName",
	"ext":          "Ext",
	"group":        "Group",
	"crc":          "CRC",
	"episodemin":   "EpisodeMin",
	"episodemax":   "EpisodeMax",
	"epname":       "EpName",
}

// convertPyFormat converts Python-style %(key)s placeholders to Go {{.Field}} syntax.
// This allows tvnamer-compatible config files to work without modification.
func convertPyFormat(s string) string {
	return pyFmtRe.ReplaceAllStringFunc(s, func(match string) string {
		sub := pyFmtRe.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		if field, ok := pyKeyField[sub[1]]; ok {
			return "{{." + field + "}}"
		}
		return match
	})
}

// EpisodeData is the data passed to output filename templates.
type EpisodeData struct {
	SeriesName       string
	SeasonNumber     string
	Episode          string
	EpisodeName      string
	Ext              string
	Group            string
	CRC              string
	OriginalFilename string
	Year             string
	Month            string
	Day              string
}

// FormatFilename generates the output filename from a template and episode data.
func FormatFilename(cfg *Config, parsed *ParseResult, seriesName string, episodeNames []string) (string, error) {
	tmplStr := selectTemplate(cfg, parsed, episodeNames)
	epStr := formatEpisodeNumbers(cfg, parsed)
	epName := formatEpisodeName(cfg, parsed, episodeNames)

	data := EpisodeData{
		SeriesName:       seriesName,
		SeasonNumber:     fmt.Sprintf("%02d", parsed.SeasonNumber),
		Episode:          epStr,
		EpisodeName:      epName,
		Ext:              parsed.Extension,
		Group:            parsed.Group,
		CRC:              parsed.CRC,
		OriginalFilename: parsed.BaseName + parsed.Extension,
	}

	if parsed.EpisodeType == EpisodeDateBased {
		data.Year = fmt.Sprintf("%04d", parsed.Year)
		data.Month = fmt.Sprintf("%02d", parsed.Month)
		data.Day = fmt.Sprintf("%02d", parsed.Day)
		data.Episode = data.Year + "-" + data.Month + "-" + data.Day
	}

	tmpl, err := template.New("filename").Parse(convertPyFormat(tmplStr))
	if err != nil {
		return "", fmt.Errorf("parsing template %q: %w", tmplStr, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return buf.String(), nil
}

// selectTemplate picks the right filename template based on episode type and available data.
func selectTemplate(cfg *Config, parsed *ParseResult, episodeNames []string) string {
	hasEpName := len(episodeNames) > 0 && episodeNames[0] != ""

	switch parsed.EpisodeType {
	case EpisodeDateBased:
		if hasEpName {
			return cfg.FilenameWithDateAndEpisode
		}
		return cfg.FilenameWithDateWithoutEpisode

	case EpisodeAnime:
		hasCRC := parsed.CRC != ""
		if hasCRC {
			if hasEpName {
				return cfg.FilenameAnimeWithEpisode
			}
			return cfg.FilenameAnimeWithoutEpisode
		}
		if hasEpName {
			return cfg.FilenameAnimeWithEpisodeWithoutCRC
		}
		return cfg.FilenameAnimeWithoutEpisodeWithoutCRC

	default: // Standard
		if parsed.HasSeason {
			if hasEpName {
				return cfg.FilenameWithEpisode
			}
			return cfg.FilenameWithoutEpisode
		}
		if hasEpName {
			return cfg.FilenameWithEpisodeNoSeason
		}
		return cfg.FilenameWithoutEpisodeNoSeason
	}
}

// formatEpisodeNumbers formats episode numbers according to config.
func formatEpisodeNumbers(cfg *Config, parsed *ParseResult) string {
	if len(parsed.EpisodeNumbers) == 0 {
		return ""
	}

	parts := make([]string, len(parsed.EpisodeNumbers))
	for i, ep := range parsed.EpisodeNumbers {
		parts[i] = fmt.Sprintf(cfg.EpisodeSingle, ep)
	}
	return strings.Join(parts, cfg.EpisodeSeparator)
}

// formatEpisodeName handles multi-episode name formatting.
func formatEpisodeName(cfg *Config, parsed *ParseResult, names []string) string {
	if len(names) == 0 {
		return ""
	}
	if len(names) == 1 {
		return names[0]
	}

	// Check if all names are the same
	allSame := true
	for _, n := range names[1:] {
		if n != names[0] {
			allSame = false
			break
		}
	}

	if allSame {
		// Use multiep_format template
		epMin := fmt.Sprintf(cfg.EpisodeSingle, parsed.EpisodeNumbers[0])
		epMax := fmt.Sprintf(cfg.EpisodeSingle, parsed.EpisodeNumbers[len(parsed.EpisodeNumbers)-1])

		tmpl, err := template.New("multiep").Parse(convertPyFormat(cfg.MultiepFormat))
		if err != nil {
			return names[0]
		}
		var buf bytes.Buffer
		data := struct {
			EpName     string
			EpisodeMin string
			EpisodeMax string
		}{
			EpName:     names[0],
			EpisodeMin: epMin,
			EpisodeMax: epMax,
		}
		if err := tmpl.Execute(&buf, data); err != nil {
			return names[0]
		}
		return buf.String()
	}

	// Different names: join with separator
	return strings.Join(names, cfg.MultiepJoinNameWith)
}
