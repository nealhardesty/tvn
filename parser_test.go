package main

import (
	"testing"
)

func TestParserStandardSxxExx(t *testing.T) {
	cfg := DefaultConfig()
	parser, err := NewParser(&cfg)
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	tests := []struct {
		name           string
		input          string
		wantSeries     string
		wantSeason     int
		wantEpisodes   []int
		wantHasSeason  bool
	}{
		{
			name:          "basic s01e01",
			input:         "brass.eye.s01e01.avi",
			wantSeries:    "Brass Eye",
			wantSeason:    1,
			wantEpisodes:  []int{1},
			wantHasSeason: true,
		},
		{
			name:          "uppercase S01E02",
			input:         "Some.Show.S01E02.720p.avi",
			wantSeries:    "Some Show",
			wantSeason:    1,
			wantEpisodes:  []int{2},
			wantHasSeason: true,
		},
		{
			name:          "multi episode s01e23e24",
			input:         "scrubs.s01e23e24.avi",
			wantSeries:    "Scrubs",
			wantSeason:    1,
			wantEpisodes:  []int{23, 24},
			wantHasSeason: true,
		},
		{
			name:          "three episodes s01e01e02e03",
			input:         "show.name.s02e01e02e03.mkv",
			wantSeries:    "Show Name",
			wantSeason:    2,
			wantEpisodes:  []int{1, 2, 3},
			wantHasSeason: true,
		},
		{
			name:          "range s01e02-04",
			input:         "show.name.s01e02-04.avi",
			wantSeries:    "Show Name",
			wantSeason:    1,
			wantEpisodes:  []int{2, 3, 4},
			wantHasSeason: true,
		},
		{
			name:          "range with E s01e02-e04",
			input:         "show.name.s01e02-e04.avi",
			wantSeries:    "Show Name",
			wantSeason:    1,
			wantEpisodes:  []int{2, 3, 4},
			wantHasSeason: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q): %v", tt.input, err)
			}
			if result.SeriesName != tt.wantSeries {
				t.Errorf("SeriesName = %q, want %q", result.SeriesName, tt.wantSeries)
			}
			if result.SeasonNumber != tt.wantSeason {
				t.Errorf("SeasonNumber = %d, want %d", result.SeasonNumber, tt.wantSeason)
			}
			if result.HasSeason != tt.wantHasSeason {
				t.Errorf("HasSeason = %v, want %v", result.HasSeason, tt.wantHasSeason)
			}
			if !intSliceEqual(result.EpisodeNumbers, tt.wantEpisodes) {
				t.Errorf("EpisodeNumbers = %v, want %v", result.EpisodeNumbers, tt.wantEpisodes)
			}
		})
	}
}

func TestParserNxNN(t *testing.T) {
	cfg := DefaultConfig()
	parser, err := NewParser(&cfg)
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	tests := []struct {
		name         string
		input        string
		wantSeries   string
		wantSeason   int
		wantEpisodes []int
	}{
		{
			name:         "1x23",
			input:        "foo.1x23.avi",
			wantSeries:   "Foo",
			wantSeason:   1,
			wantEpisodes: []int{23},
		},
		{
			name:         "1x23x24 multi",
			input:        "foo.1x23x24.avi",
			wantSeries:   "Foo",
			wantSeason:   1,
			wantEpisodes: []int{23, 24},
		},
		{
			name:         "1x23-24 range",
			input:        "foo.1x23-24.avi",
			wantSeries:   "Foo",
			wantSeason:   1,
			wantEpisodes: []int{23, 24},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q): %v", tt.input, err)
			}
			if result.SeriesName != tt.wantSeries {
				t.Errorf("SeriesName = %q, want %q", result.SeriesName, tt.wantSeries)
			}
			if result.SeasonNumber != tt.wantSeason {
				t.Errorf("SeasonNumber = %d, want %d", result.SeasonNumber, tt.wantSeason)
			}
			if !intSliceEqual(result.EpisodeNumbers, tt.wantEpisodes) {
				t.Errorf("EpisodeNumbers = %v, want %v", result.EpisodeNumbers, tt.wantEpisodes)
			}
		})
	}
}

func TestParserAnime(t *testing.T) {
	cfg := DefaultConfig()
	parser, err := NewParser(&cfg)
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	tests := []struct {
		name         string
		input        string
		wantSeries   string
		wantEpisodes []int
		wantGroup    string
		wantCRC      string
	}{
		{
			name:         "anime with CRC",
			input:        "[Shinsen-Subs] Beet - 19 [24DAB497].mkv",
			wantSeries:   "Beet",
			wantEpisodes: []int{19},
			wantGroup:    "Shinsen-Subs",
			wantCRC:      "24DAB497",
		},
		{
			name:         "anime without CRC",
			input:        "[SubGroup] My Show - 05.mkv",
			wantSeries:   "My Show",
			wantEpisodes: []int{5},
			wantGroup:    "SubGroup",
			wantCRC:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q): %v", tt.input, err)
			}
			if result.SeriesName != tt.wantSeries {
				t.Errorf("SeriesName = %q, want %q", result.SeriesName, tt.wantSeries)
			}
			if !intSliceEqual(result.EpisodeNumbers, tt.wantEpisodes) {
				t.Errorf("EpisodeNumbers = %v, want %v", result.EpisodeNumbers, tt.wantEpisodes)
			}
			if result.Group != tt.wantGroup {
				t.Errorf("Group = %q, want %q", result.Group, tt.wantGroup)
			}
			if result.CRC != tt.wantCRC {
				t.Errorf("CRC = %q, want %q", result.CRC, tt.wantCRC)
			}
			if result.EpisodeType != EpisodeAnime {
				t.Errorf("EpisodeType = %v, want EpisodeAnime", result.EpisodeType)
			}
		})
	}
}

func TestParserDateBased(t *testing.T) {
	cfg := DefaultConfig()
	parser, err := NewParser(&cfg)
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	result, err := parser.Parse("show.2010.01.02.stuff.avi")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if result.SeriesName != "Show" {
		t.Errorf("SeriesName = %q, want %q", result.SeriesName, "Show")
	}
	if result.Year != 2010 || result.Month != 1 || result.Day != 2 {
		t.Errorf("Date = %d-%d-%d, want 2010-1-2", result.Year, result.Month, result.Day)
	}
	if result.EpisodeType != EpisodeDateBased {
		t.Errorf("EpisodeType = %v, want EpisodeDateBased", result.EpisodeType)
	}
}

func TestParserBracketed(t *testing.T) {
	cfg := DefaultConfig()
	parser, err := NewParser(&cfg)
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	tests := []struct {
		name         string
		input        string
		wantSeries   string
		wantSeason   int
		wantEpisodes []int
		wantHasSeason bool
	}{
		{
			name:          "bracketed 1x09",
			input:         "foo.[1x09].avi",
			wantSeries:    "Foo",
			wantSeason:    1,
			wantEpisodes:  []int{9},
			wantHasSeason: true,
		},
		{
			name:          "bracketed range 1x09-11",
			input:         "foo.[1x09-11].avi",
			wantSeries:    "Foo",
			wantSeason:    1,
			wantEpisodes:  []int{9, 10, 11},
			wantHasSeason: true,
		},
		{
			name:          "bracketed episode only",
			input:         "foo - [012].avi",
			wantSeries:    "Foo",
			wantSeason:    0,
			wantEpisodes:  []int{12},
			wantHasSeason: false,
		},
		{
			name:          "bracketed season.episode",
			input:         "foo - [01.09].avi",
			wantSeries:    "Foo",
			wantSeason:    1,
			wantEpisodes:  []int{9},
			wantHasSeason: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q): %v", tt.input, err)
			}
			if result.SeriesName != tt.wantSeries {
				t.Errorf("SeriesName = %q, want %q", result.SeriesName, tt.wantSeries)
			}
			if result.SeasonNumber != tt.wantSeason {
				t.Errorf("SeasonNumber = %d, want %d", result.SeasonNumber, tt.wantSeason)
			}
			if result.HasSeason != tt.wantHasSeason {
				t.Errorf("HasSeason = %v, want %v", result.HasSeason, tt.wantHasSeason)
			}
			if !intSliceEqual(result.EpisodeNumbers, tt.wantEpisodes) {
				t.Errorf("EpisodeNumbers = %v, want %v", result.EpisodeNumbers, tt.wantEpisodes)
			}
		})
	}
}

func TestParserMiscFormats(t *testing.T) {
	cfg := DefaultConfig()
	parser, err := NewParser(&cfg)
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	tests := []struct {
		name         string
		input        string
		wantSeries   string
		wantSeason   int
		wantEpisodes []int
	}{
		{
			name:         "Season Episode spelled out",
			input:        "show name Season 01 Episode 20.avi",
			wantSeries:   "Show Name",
			wantSeason:   1,
			wantEpisodes: []int{20},
		},
		{
			name:         "compact s0101",
			input:        "foo.s0101.avi",
			wantSeries:   "Foo",
			wantSeason:   1,
			wantEpisodes: []int{1},
		},
		{
			name:         "part notation",
			input:        "Show.Name.Part1.avi",
			wantSeries:   "Show Name",
			wantSeason:   0,
			wantEpisodes: []int{1},
		},
		{
			name:         "part and part",
			input:        "Show.Name.Part.1.and.Part.2.avi",
			wantSeries:   "Show Name",
			wantSeason:   0,
			wantEpisodes: []int{1, 2},
		},
		{
			name:         "N of M",
			input:        "show name 2 of 6.avi",
			wantSeries:   "Show Name",
			wantSeason:   0,
			wantEpisodes: []int{2},
		},
		{
			name:         "simple e format",
			input:        "show.name.e123.avi",
			wantSeries:   "Show Name",
			wantSeason:   0,
			wantEpisodes: []int{123},
		},
		{
			name:         "alternate S2 E 02",
			input:        "Foo - S2 E 02 - stuff.avi",
			wantSeries:   "Foo",
			wantSeason:   2,
			wantEpisodes: []int{2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q): %v", tt.input, err)
			}
			if result.SeriesName != tt.wantSeries {
				t.Errorf("SeriesName = %q, want %q", result.SeriesName, tt.wantSeries)
			}
			if result.SeasonNumber != tt.wantSeason {
				t.Errorf("SeasonNumber = %d, want %d", result.SeasonNumber, tt.wantSeason)
			}
			if !intSliceEqual(result.EpisodeNumbers, tt.wantEpisodes) {
				t.Errorf("EpisodeNumbers = %v, want %v", result.EpisodeNumbers, tt.wantEpisodes)
			}
		})
	}
}

func TestParserExtensionSplit(t *testing.T) {
	cfg := DefaultConfig()
	parser, err := NewParser(&cfg)
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	result, err := parser.Parse("show.s01e01.avi")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if result.Extension != ".avi" {
		t.Errorf("Extension = %q, want %q", result.Extension, ".avi")
	}

	result, err = parser.Parse("show.s01e01.720p.mkv")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if result.Extension != ".mkv" {
		t.Errorf("Extension = %q, want %q", result.Extension, ".mkv")
	}
}

func TestParserNoMatch(t *testing.T) {
	cfg := DefaultConfig()
	parser, err := NewParser(&cfg)
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	_, err = parser.Parse("random_file.txt")
	if err == nil {
		t.Error("expected error for non-matching filename")
	}
}

func intSliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
