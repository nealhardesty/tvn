package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Replacement defines a string or regex replacement rule.
type Replacement struct {
	IsRegex     bool   `json:"is_regex"`
	Match       string `json:"match"`
	Replacement string `json:"replacement"`
}

// Config holds all tvnamer configuration.
type Config struct {
	// Batch / Behavior
	SelectFirst                  bool   `json:"select_first"`
	AlwaysRename                 bool   `json:"always_rename"`
	Batch                        bool   `json:"batch"`
	SkipFileOnError              bool   `json:"skip_file_on_error"`
	SkipBehaviour                string `json:"skip_behaviour"`
	OverwriteDestinationOnRename bool   `json:"overwrite_destination_on_rename"`
	OverwriteDestinationOnMove   bool   `json:"overwrite_destination_on_move"`

	// Output / Debug
	Verbose bool `json:"verbose"`
	DryRun  bool `json:"dry_run"`

	// File Discovery
	Recursive         bool     `json:"recursive"`
	ValidExtensions   []string `json:"valid_extensions"`
	ExtensionPattern  string   `json:"extension_pattern"`
	FilenameBlacklist []string `json:"filename_blacklist"`

	// Filename Sanitization
	WindowsSafeFilenames             bool   `json:"windows_safe_filenames"`
	NormalizeUnicodeFilenames        bool   `json:"normalize_unicode_filenames"`
	LowercaseFilename               bool   `json:"lowercase_filename"`
	TitlecaseFilename                bool   `json:"titlecase_filename"`
	CustomFilenameCharacterBlacklist string `json:"custom_filename_character_blacklist"`
	ReplaceInvalidCharactersWith     string `json:"replace_invalid_characters_with"`

	// Replacements
	InputFilenameReplacements     []Replacement `json:"input_filename_replacements"`
	OutputFilenameReplacements    []Replacement `json:"output_filename_replacements"`
	MoveFilesFullpathReplacements []Replacement `json:"move_files_fullpath_replacements"`

	// TheTVDB / Metadata
	TVDBApiKey              *string           `json:"tvdb_api_key"`
	Language                string            `json:"language"`
	SearchAllLanguages      bool              `json:"search_all_languages"`
	SeriesID                *int              `json:"series_id"`
	ForceName               *string           `json:"force_name"`
	InputSeriesReplacements  map[string]string `json:"input_series_replacements"`
	OutputSeriesReplacements map[string]string `json:"output_series_replacements"`
	Order                   string            `json:"order"`

	// Move Files
	MoveFilesEnable               bool   `json:"move_files_enable"`
	MoveFilesConfirmation         bool   `json:"move_files_confirmation"`
	MoveFilesLowercaseDestination bool   `json:"move_files_lowercase_destination"`
	MoveFilesDestinationIsFilepath bool   `json:"move_files_destination_is_filepath"`
	MoveFilesDestination          string `json:"move_files_destination"`
	MoveFilesDestinationDate      string `json:"move_files_destination_date"`
	AlwaysMove                    bool   `json:"always_move"`
	LeaveSymlink                  bool   `json:"leave_symlink"`
	MoveFilesOnly                 bool   `json:"move_files_only"`

	// Input Parsing Patterns (if empty, defaults are used)
	FilenamePatterns []string `json:"filename_patterns"`

	// Output Filename Formats
	FilenameWithEpisode                   string `json:"filename_with_episode"`
	FilenameWithoutEpisode                string `json:"filename_without_episode"`
	FilenameWithEpisodeNoSeason           string `json:"filename_with_episode_no_season"`
	FilenameWithoutEpisodeNoSeason        string `json:"filename_without_episode_no_season"`
	FilenameWithDateAndEpisode            string `json:"filename_with_date_and_episode"`
	FilenameWithDateWithoutEpisode        string `json:"filename_with_date_without_episode"`
	FilenameAnimeWithEpisode              string `json:"filename_anime_with_episode"`
	FilenameAnimeWithoutEpisode           string `json:"filename_anime_without_episode"`
	FilenameAnimeWithEpisodeWithoutCRC    string `json:"filename_anime_with_episode_without_crc"`
	FilenameAnimeWithoutEpisodeWithoutCRC string `json:"filename_anime_without_episode_without_crc"`

	// Episode Formatting
	EpisodeSingle       string `json:"episode_single"`
	EpisodeSeparator    string `json:"episode_separator"`
	MultiepJoinNameWith string `json:"multiep_join_name_with"`
	MultiepFormat       string `json:"multiep_format"`
}

// DefaultConfig returns a Config with all default values.
func DefaultConfig() Config {
	return Config{
		SkipFileOnError:              true,
		SkipBehaviour:                "skip",
		ReplaceInvalidCharactersWith: "",
		ExtensionPattern:             `(\.[a-zA-Z0-9]+)$`,
		Language:                     "en",
		SearchAllLanguages:           true,
		NormalizeUnicodeFilenames:    true,
		Order:                        "aired",

		MoveFilesConfirmation:    true,
		MoveFilesDestination:     ".",
		MoveFilesDestinationDate: ".",

		FilenameWithEpisode:                   `{{.SeriesName}}.S{{.SeasonNumber}}E{{.Episode}}.{{.EpisodeName}}{{.Ext}}`,
		FilenameWithoutEpisode:                `{{.SeriesName}}.{{.SeasonNumber}}x{{.Episode}}{{.Ext}}`,
		FilenameWithEpisodeNoSeason:           `{{.SeriesName}}.E{{.Episode}}.{{.EpisodeName}}{{.Ext}}`,
		FilenameWithoutEpisodeNoSeason:        `{{.SeriesName}}.{{.Episode}}{{.Ext}}`,
		FilenameWithDateAndEpisode:            `{{.SeriesName}}.{{.Episode}}.{{.EpisodeName}}{{.Ext}}`,
		FilenameWithDateWithoutEpisode:        `{{.SeriesName}}.{{.Episode}}{{.Ext}}`,
		FilenameAnimeWithEpisode:              `[{{.Group}}] {{.SeriesName}} - {{.Episode}} - {{.EpisodeName}} [{{.CRC}}]{{.Ext}}`,
		FilenameAnimeWithoutEpisode:           `[{{.Group}}] {{.SeriesName}} - {{.Episode}} [{{.CRC}}]{{.Ext}}`,
		FilenameAnimeWithEpisodeWithoutCRC:    `[{{.Group}}] {{.SeriesName}} - {{.Episode}} - {{.EpisodeName}}{{.Ext}}`,
		FilenameAnimeWithoutEpisodeWithoutCRC: `[{{.Group}}] {{.SeriesName}} - {{.Episode}}{{.Ext}}`,

		EpisodeSingle:       "%02d",
		EpisodeSeparator:    "E",
		MultiepJoinNameWith: ", ",
		MultiepFormat:       `({{.EpisodeMin}}-{{.EpisodeMax}}).{{.EpName}}`,

		ValidExtensions:               []string{},
		FilenameBlacklist:             []string{},
		InputFilenameReplacements:     []Replacement{},
		OutputFilenameReplacements:    []Replacement{
			{IsRegex: true, Match: `[ ]`, Replacement: "."},
			{IsRegex: true, Match: `[':?!@#$%^&*]`, Replacement: ""},
			{IsRegex: true, Match: `[&]`, Replacement: "and"},
		},
		MoveFilesFullpathReplacements: []Replacement{},
		InputSeriesReplacements:       map[string]string{},
		OutputSeriesReplacements:      map[string]string{},
	}
}

// FindConfigFile locates the config file on disk.
func FindConfigFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	xdg := filepath.Join(home, ".config", "tvn", "tvn.json")
	if _, err := os.Stat(xdg); err == nil {
		return xdg
	}
	legacy := filepath.Join(home, ".tvn.json")
	if _, err := os.Stat(legacy); err == nil {
		return legacy
	}
	return ""
}

// LoadConfig loads a config from a JSON file, merging onto defaults.
func LoadConfig(path string) (Config, error) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("reading config: %w", err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing config: %w", err)
	}
	return cfg, nil
}

// SaveConfig writes the config as formatted JSON to a file.
func SaveConfig(cfg Config, path string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}
