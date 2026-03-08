# Changelog

## [0.1.15] - 2026-03-07

### Fixed
- Strip year-in-parens (e.g. `(2025)`) from parsed series name before TVDB search, preventing garbage matches when filenames include the year
- Always prompt for series selection unless `--select-first` or `--batch` is set; previously a single TVDB result was silently auto-selected, which could pick the wrong show without any user confirmation

## [0.1.12] - 2026-03-07

### Changed
- Updated `DefaultConfig()` to match user-preferred defaults: episode separator `"E"`, normalize unicode filenames enabled, `replace_invalid_characters_with` empty string, default `output_filename_replacements` (spaces→dots, strip special chars, `&`→`and`), and dot-separated filename templates (`SeriesName.S01E01.EpisodeName`)
- Overhauled README.md: corrected install command module path, updated example output format to reflect new defaults, documented all CLI flags (including `--not-*` variants), added full configuration reference table, documented template variables, move file options, episode formatting, and `TVN_TEST_MODE` environment variable

## [0.1.9] - 2026-03-07

### Fixed
- "Always rename" (selecting 'a') is now scoped per-series; it no longer carries over to files from a different series

## [0.1.8] - 2026-03-07

### Changed
- TVDB search results header now shows the source filename in parens for context

## [0.1.7] - 2026-03-07

### Changed
- Print "<filename> is already correct" when no rename is needed (previously only shown with --verbose)

## [0.1.6] - 2026-03-07

### Changed
- All interactive prompts now accept Enter as a default: series selection defaults to item 1, rename and move prompts default to yes (`[Y/n/...]`)

## [0.1.5] - 2026-03-07

### Fixed
- Filename templates from tvnamer-style configs using Python `%(key)s` placeholders now render correctly. The formatter converts them to Go template syntax before execution, so existing tvnamer config files work without modification.

## [0.1.4] - 2026-03-07

### Changed
- Cap series selection list to 10 items when TVDB returns more results

## [0.1.3] - 2026-03-06

### Changed
- Bumped patch version

## [0.1.2] - 2026-03-06

### Changed
- Bumped patch version

## [0.1.1] - 2026-03-06

### Changed
- Renamed project to 'tvn'
- Refactored version management into version.go

## [0.1.0] - 2026-03-06

### Added
- Initial release
- TVDB series search and episode renaming
- Default TVDB API key for easier setup
