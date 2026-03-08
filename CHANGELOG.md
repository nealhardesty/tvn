# Changelog

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
