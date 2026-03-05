# Product Requirements Document: tvnamer (Go)

## 1. Overview

### 1.1 Product Summary

**tvn** is a command-line utility that automatically renames TV episode files from common download formats (e.g., `some.show.s01e03.blah.abc.avi`) to human-readable, consistent formats (e.g., `Some Show - [01x03] - The Episode Name.avi`) by retrieving episode metadata from [TheTVDB.com](https://thetvdb.com).

This is a Go reimplementation of the original Python [tvnamer](https://github.com/dbr/tvnamer) project.

**License:** Unlicense

**Note:** TV information is provided by TheTVDB.com. tvn is not endorsed or certified by TheTVDB.com or its affiliates.

---

### 1.2 Target Users

- Media enthusiasts organizing TV episode collections
- Users who download or record TV episodes and want consistent naming
- Users with anime collections using subtitle group naming conventions
- Power users automating media organization workflows

---

### 1.3 Supported Platforms

- **Language:** Go 1.22+
- **Operating Systems:** OS-independent (macOS, Linux, Windows, FreeBSD)
- **Distribution:** Single static binary per platform; `go install`, Homebrew, package managers, GitHub releases

---

## 2. Core Capabilities

### 2.1 File Renaming

| Capability | Description |
|------------|-------------|
| **Parse filenames** | Extracts series name, season, episode number(s), and optional metadata (e.g., CRC for anime) from input filenames using Go `regexp` |
| **Fetch metadata** | Queries TheTVDB API (v4) to retrieve official series name and episode titles |
| **Generate output names** | Applies configurable output format templates using Go `fmt.Sprintf`-style or Go `text/template` formatting |
| **Rename on disk** | Renames files in place or moves them to a configurable destination using `os.Rename` / file copy for cross-device moves |

### 2.2 Input Filename Formats Supported

tvn supports a comprehensive set of filename patterns, including:

| Format Type | Example |
|-------------|---------|
| **Standard SxxExx** | `brass.eye.s01e01.avi`, `scrubs.s01e23e24.avi` |
| **Season x Episode** | `foo.1x23.avi`, `foo.1x23x24.avi`, `foo.s01e23-24.avi`, `foo.1x23-24.avi` |
| **Bracketed** | `foo.[1x09-11].avi`, `foo - [012].avi`, `foo - [01.09].avi` |
| **Compact SxxExx** | `foo.s0101.avi`, `foo.0201.avi` |
| **Anime** | `[Shinsen-Subs] Beet - 19 [24DAB497].mkv`, `[Group] Series - 02 - Episode Name [CRC].ext` |
| **Date-based** | `show.2010.01.02.etc` |
| **Alternate formats** | `Foo - S2 E 02 - etc`, `Show - Episode 9999 [S 12 - Ep 131] - etc` |
| **Part notation** | `show name 2 of 6`, `Show.Name.Part.1.and.Part.2`, `Show.Name.Part1` |
| **Season/Episode spelled out** | `show name Season 01 Episode 20` |
| **Simple E-format** | `show.name.e123.abc` |

### 2.3 Multi-Episode Files

- Supports files containing multiple episodes (e.g., `scrubs.s01e23e24.avi`)
- Supports consecutive episode ranges (e.g., `episodenumberstart` and `episodenumberend`)
- Configurable episode separator (default `-`) for multi-episode output
- Configurable multi-episode name format when episode names differ or are the same

---

## 3. User Interface

### 3.1 Interaction Modes

| Mode | Flags | Behavior |
|------|-------|----------|
| **Interactive** | (default) | Prompts for series selection when multiple matches; prompts for each rename |
| **Always** | `-a`, `--always` | Auto-renames after series selection; still prompts for series choice |
| **Select First** | `-f`, `--selectfirst` | Auto-selects first search result; still prompts for rename |
| **Batch** | `-b`, `--batch` | No prompts; auto-selects first series and renames all files |

### 3.2 Interactive Prompts

- **Series selection:** When multiple language/region variants exist (e.g., "Lost" in en, sv, no, fi, nl, de), user selects by number
- **Rename confirmation:** For each file: `(y/n/a/q)` — yes, no, always (for rest), quit
- **Move confirmation:** When moving files, optional separate prompt (`y/n/q`)

---

## 4. Command-Line Interface

### 4.1 Usage

```
tvn [options] <files or directories>
```

CLI argument parsing should use a Go flag library (e.g., `pflag`, `cobra`, or the standard `flag` package with POSIX-style support).

### 4.2 Command-Line Arguments

#### Console Output

| Argument | Description |
|----------|-------------|
| `-v`, `--verbose` | Show debugging info |
| `-q`, `--not-verbose` | Disable verbose (overrides config) |
| `--dry-run` | Show planned actions without renaming or moving |

#### Batch Options

| Argument | Description |
|----------|-------------|
| `-a`, `--always` | Always rename after series selection |
| `--not-always` | Override `--always` |
| `-f`, `--selectfirst` | Select first series search result automatically |
| `--not-selectfirst` | Override `--selectfirst` |
| `-b`, `--batch` | No prompts; same as `--always` + `--selectfirst` |
| `--not-batch` | Override `--batch` |

#### Config Options

| Argument | Description |
|----------|-------------|
| `-c`, `--config <path>` | Load config from specified file |
| `-s`, `--save <path>` | Save current config to file and exit |
| `-p`, `--preview-config` | Show current config values and exit |

#### Override Values

| Argument | Description |
|----------|-------------|
| `-n`, `--name <name>` | Override parsed series name for all files |
| `--series-id <id>` | Use specific TheTVDB series ID instead of searching |
| `--order <aired\|dvd>` | Episode order: `aired` (default) or `dvd` |
| `-l`, `--lang <code>` | Language for metadata (e.g., `de`, `fr`) |

#### Miscellaneous

| Argument | Description |
|----------|-------------|
| `-r`, `--recursive` | Descend more than one level into directories |
| `--not-recursive` | Only one level into directories (default) |
| `-m`, `--move` | Move files to configured destination |
| `--not-move` | Do not move files (default) |
| `--force-move` | Overwrite existing files in destination on move |
| `--force-rename` | Overwrite destination when renaming |
| `-d`, `--movedestination <path>` | Override move destination (supports format variables) |
| `-h`, `--help` | Show help |
| `--version` | Show version |

---

## 5. Configuration

### 5.1 Config File Locations

| Location | Priority |
|----------|----------|
| `--config <path>` | Highest (CLI override) |
| `~/.config/tvn/tvn.json` | Default |
| `~/.tvn.json` | Legacy (deprecated) |

Config is JSON. Deserialized into a Go struct with `json` tags. Unknown fields should be ignored (forward compatibility).

### 5.2 Generate Default Config

```bash
tvn --save=./mytvnconfig.json
```

### 5.3 Configuration Reference

#### Batch / Behavior

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `select_first` | bool | false | Auto-select first series result |
| `always_rename` | bool | false | Auto-rename after series selection |
| `batch` | bool | false | Same as select_first + always_rename |
| `skip_file_on_error` | bool | true | Skip file on metadata error (when always_rename) |
| `skip_behaviour` | `"skip"` \| `"error"` | `"skip"` | `skip` = skip file; `error` = exit on error |
| `overwrite_destination_on_rename` | bool | false | Overwrite existing file when renaming |
| `overwrite_destination_on_move` | bool | false | Overwrite existing file when moving |

#### Output / Debug

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `verbose` | bool | false | Debug output |
| `dry_run` | bool | false | Preview only, no changes |

#### File Discovery

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `recursive` | bool | false | Recurse into subdirectories |
| `valid_extensions` | []string | `[]` | Restrict to extensions (e.g., `["avi","mkv","mp4"]`) |
| `extension_pattern` | string | `(\.[a-zA-Z0-9]+)$` | Regex for basename/extension split (e.g., for `.eng.srt`) |
| `filename_blacklist` | []string | `[]` | Exclude files matching these (simple or regex) |

#### Filename Sanitization

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `windows_safe_filenames` | bool | false | Force Windows-safe output |
| `normalize_unicode_filenames` | bool | false | Replace accented chars with ASCII (use `golang.org/x/text/unicode/norm`) |
| `lowercase_filename` | bool | false | Lowercase output |
| `titlecase_filename` | bool | false | Title case output (use `golang.org/x/text/cases`) |
| `custom_filename_character_blacklist` | string | `""` | Extra invalid chars |
| `replace_invalid_characters_with` | string | `"_"` | Replacement for invalid chars |

#### Replacements

| Option | Type | Description |
|--------|------|-------------|
| `input_filename_replacements` | []Replacement | Applied before parsing; `[{"is_regex": bool, "match": string, "replacement": string}]` |
| `output_filename_replacements` | []Replacement | Applied after generating new name |
| `move_files_fullpath_replacements` | []Replacement | Applied to full move path |

#### TheTVDB / Metadata

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `tvdb_api_key` | string \| null | null | Custom TheTVDB API key |
| `language` | string | `"en"` | Metadata language (2-letter code) |
| `search_all_languages` | bool | true | Return multilingual results |
| `series_id` | int \| null | null | Force TheTVDB series ID |
| `force_name` | string \| null | null | Force series name |
| `input_series_replacements` | map[string]string | `{}` | Regex replacements for parsed series name |
| `output_series_replacements` | map[string]string | `{}` | String replacements for TVDB series name |
| `order` | `"aired"` \| `"dvd"` | `"aired"` | Episode order |

#### Move Files

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `move_files_enable` | bool | false | Enable move after rename |
| `move_files_confirmation` | bool | true | Prompt before moving |
| `move_files_lowercase_destination` | bool | false | Lowercase dynamic path parts |
| `move_files_destination_is_filepath` | bool | false | Destination includes filename |
| `move_files_destination` | string | `"."` | Path template; vars: `{{.SeriesName}}`, `{{.SeasonNumber}}`, `{{.EpisodeNumbers}}`, `{{.OriginalFilename}}` |
| `move_files_destination_date` | string | `"."` | For date-based episodes; vars: `{{.SeriesName}}`, `{{.Year}}`, `{{.Month}}`, `{{.Day}}`, `{{.OriginalFilename}}` |
| `always_move` | bool | false | Delete source when copy+move (e.g., cross-partition) |
| `leave_symlink` | bool | false | Leave symlink to new location |
| `move_files_only` | bool | false | Move without renaming |

#### Input Parsing Patterns

| Option | Type | Description |
|--------|------|-------------|
| `filename_patterns` | []string | Go `regexp` patterns with named groups: `seriesname`, `seasonnumber` (optional), `episodenumber` or `episodenumber1`/`episodenumber2` or `episodenumberstart`/`episodenumberend` |

**Note:** Go's `regexp` package uses RE2 syntax. Named groups use `(?P<name>...)` syntax. Backreferences are not supported. Patterns from the original Python version that use features outside RE2 must be adapted.

#### Output Filename Formats

Go `fmt.Sprintf`-style formatting. Variables are interpolated via struct fields. The format strings use `{{.Field}}` Go template syntax.

| Option | Default |
|--------|---------|
| `filename_with_episode` | `{{.SeriesName}} - [{{.SeasonNumber}}x{{.Episode}}] - {{.EpisodeName}}{{.Ext}}` |
| `filename_without_episode` | `{{.SeriesName}} - [{{.SeasonNumber}}x{{.Episode}}]{{.Ext}}` |
| `filename_with_episode_no_season` | `{{.SeriesName}} - [{{.Episode}}] - {{.EpisodeName}}{{.Ext}}` |
| `filename_without_episode_no_season` | `{{.SeriesName}} - [{{.Episode}}]{{.Ext}}` |
| `filename_with_date_and_episode` | `{{.SeriesName}} - [{{.Episode}}] - {{.EpisodeName}}{{.Ext}}` |
| `filename_with_date_without_episode` | `{{.SeriesName}} - [{{.Episode}}]{{.Ext}}` |
| `filename_anime_with_episode` | `[{{.Group}}] {{.SeriesName}} - {{.Episode}} - {{.EpisodeName}} [{{.CRC}}]{{.Ext}}` |
| `filename_anime_without_episode` | `[{{.Group}}] {{.SeriesName}} - {{.Episode}} [{{.CRC}}]{{.Ext}}` |
| `filename_anime_with_episode_without_crc` | `[{{.Group}}] {{.SeriesName}} - {{.Episode}} - {{.EpisodeName}}{{.Ext}}` |
| `filename_anime_without_episode_without_crc` | `[{{.Group}}] {{.SeriesName}} - {{.Episode}}{{.Ext}}` |

**Note:** Season numbers should be zero-padded to 2 digits by default (e.g., `{{printf "%02d" .SeasonNumber}}`). Episode numbers use configurable formatting.

#### Episode Formatting

| Option | Default | Description |
|--------|---------|-------------|
| `episode_single` | `"%02d"` | `fmt.Sprintf` format for single episode number |
| `episode_separator` | `"-"` | Joiner for multi-episode numbers |
| `multiep_join_name_with` | `", "` | Joiner when episode names differ |
| `multiep_format` | `{{.EpName}} ({{.EpisodeMin}}-{{.EpisodeMax}})` | Format when names are same |

---

## 6. File Operations

### 6.1 Rename Behavior

- Renames in place by default using `os.Rename`
- `overwrite_destination_on_rename`: overwrite if target exists
- Cross-device moves handled by copy + remove (since `os.Rename` fails across filesystems)

### 6.2 Move Behavior

- Move destination supports template variables (see Move Files config above)
- Date-based: `{{.SeriesName}}`, `{{.Year}}`, `{{.Month}}`, `{{.Day}}`
- `move_files_destination_is_filepath`: destination is full path including filename
- `always_move`: delete source after copy when moving across partitions
- `leave_symlink`: leave symlink at original path using `os.Symlink`

### 6.3 Error Handling

- `skip_file_on_error` + `skip_behaviour: "skip"`: skip problematic file
- `skip_behaviour: "error"`: exit on first error
- User abort (`q`): exit gracefully

---

## 7. Custom Patterns and Replacements

### 7.1 Custom Input Patterns

Go `regexp` (RE2) patterns must include named groups:
- **Required:** `seriesname`
- **Optional:** `seasonnumber` (defaults to 1)
- **Episode:** `episodenumber` (single) or `episodenumber1`, `episodenumber2`, ... or `episodenumberstart`, `episodenumberend`

Named groups use Go syntax: `(?P<seriesname>...)`.

### 7.2 Regex Flags in Config

Use `(?i)` for case-insensitive (e.g., `(?i)and`). Go RE2 supports `(?i)`, `(?m)`, `(?s)`, `(?U)` flags.

---

## 8. Dependencies

| Dependency | Purpose |
|------------|---------|
| Go standard library (`regexp`, `os`, `encoding/json`, `text/template`, `net/http`) | Core functionality |
| `golang.org/x/text` | Unicode normalization and title casing |
| TheTVDB API v4 (REST/JSON) | Episode metadata (direct HTTP client, no wrapper library required) |

### 8.1 TheTVDB API

The Go version calls TheTVDB's REST API directly using `net/http` and `encoding/json` rather than relying on a wrapper library. Authentication uses bearer tokens obtained via the `/login` endpoint.

---

## 9. Technical Notes

### 9.1 Explicitly Unsupported Patterns (v4+)

- `show.123.avi` (ambiguous with H.264 etc.)
- `show.0123.avi`

### 9.2 API Key

- Default key provided for tvn
- Custom key via `tvdb_api_key` (register at [TheTVDB](https://thetvdb.com/api-information))

### 9.3 Test Mode

- `TVN_TEST_MODE=1` uses cached test data for CI (HTTP responses cached as fixtures)

### 9.4 Go-Specific Design Notes

- **Single binary:** Compile to a single static binary per platform. No runtime dependencies.
- **Concurrency:** File processing can leverage goroutines for parallel API lookups, but rename operations should be sequential to preserve predictable output ordering.
- **Config struct:** Configuration is a single Go struct with `json` tags matching the JSON field names. Use pointer types (`*bool`, `*string`) where null/absent must be distinguished from zero values.
- **RE2 regex:** Go's `regexp` uses RE2 which does not support lookaheads, lookbehinds, or backreferences. Any original Python patterns using these features must be rewritten or handled with alternative logic.
- **Cross-platform paths:** Use `filepath` package for all path operations to ensure Windows compatibility.

---

## 10. Appendix: Example Config Snippets

### Change Output Format

```json
{
  "filename_with_episode": "{{.SeriesName}} {{printf \"%02d\" .SeasonNumber}}x{{.Episode}} {{.EpisodeName}}{{.Ext}}"
}
```

### Replace Spaces with Dots

```json
{
  "output_filename_replacements": [
    {"is_regex": true, "match": "[ ]", "replacement": "."}
  ]
}
```

### Move to Series/Season Structure

```json
{
  "move_files_enable": true,
  "move_files_destination": "/media/tv/{{.SeriesName}}/Season {{.SeasonNumber}}/"
}
```

### French Metadata

```json
{
  "language": "fr"
}
```

### Case-Insensitive Replacement

```json
{
  "input_filename_replacements": [
    {"is_regex": true, "match": "(?i)and", "replacement": "&"}
  ]
}
```

---

*Go reimplementation of [tvnamer](https://github.com/dbr/tvnamer). Last updated: March 2026.*
