# tvn

A command-line utility that automatically renames TV episode files by fetching metadata from [TheTVDB.com](https://thetvdb.com). Transforms messy download filenames like `some.show.s01e03.blah.abc.avi` into clean, consistent names like `Some.Show.S01E03.The.Episode.Name.avi`.

Go reimplementation of the Python [tvnamer](https://github.com/dbr/tvnamer) project. Config files from the original tvnamer (using `%(key)s` placeholders) are supported without modification.

**Note:** TV information is provided by TheTVDB.com. tvn is not endorsed or certified by TheTVDB.com or its affiliates.

## Installation

Requires Go 1.22+.

```bash
go install github.com/nealhardesty/tvn@latest
```

Or build from source:

```bash
git clone https://github.com/nealhardesty/tvn.git
cd tvn
make install
```

## Usage

```
tvn [options] <files or directories>
```

### Examples

```bash
# Rename a single file (interactive prompts)
tvn some.show.s01e03.blah.avi

# Rename all files in a directory
tvn /path/to/episodes/

# Batch mode (no prompts)
tvn -b *.mkv

# Dry run to preview changes
tvn --dry-run *.avi

# Override the series name
tvn -n "The Office" *.mkv

# Use a specific TheTVDB series ID
tvn --series-id 73244 *.mkv

# Recursive directory processing
tvn -r /path/to/tv/

# Move files to an organized directory structure after renaming
tvn -m -d "/media/tv/{{.SeriesName}}/Season {{.SeasonNumber}}/" *.mkv

# Save current config (with all defaults) to a file
tvn --save=~/.config/tvn/tvn.json

# Preview resolved config without processing any files
tvn --preview-config
```

### Interaction Modes

| Mode | Flag | Behavior |
|------|------|----------|
| Interactive | (default) | Prompts for series selection and each rename |
| Always | `-a` | Auto-renames after series selection |
| Select First | `-f` | Auto-selects first search result |
| Batch | `-b` | No prompts at all (`-a` + `-f`) |

### Options

#### Output

| Flag | Description |
|------|-------------|
| `-v`, `--verbose` | Show debugging info |
| `-q`, `--not-verbose` | Disable verbose output |
| `--dry-run` | Preview changes without renaming |

#### Batch

| Flag | Description |
|------|-------------|
| `-a`, `--always` | Always rename after series selection |
| `--not-always` | Override `--always` from config |
| `-f`, `--selectfirst` | Auto-select first search result |
| `--not-selectfirst` | Override `--selectfirst` from config |
| `-b`, `--batch` | No prompts (`--always` + `--selectfirst`) |
| `--not-batch` | Override `--batch` from config |

#### Overrides

| Flag | Description |
|------|-------------|
| `-n`, `--name <name>` | Override parsed series name |
| `--series-id <id>` | Use specific TheTVDB series ID |
| `--order <aired\|dvd>` | Episode order (default: aired) |
| `-l`, `--lang <code>` | Metadata language (e.g., `de`, `fr`) |

#### File Handling

| Flag | Description |
|------|-------------|
| `-r`, `--recursive` | Recurse into subdirectories |
| `--not-recursive` | Disable recursive from config |
| `-m`, `--move` | Move files to destination after rename |
| `--not-move` | Disable move from config |
| `-d`, `--movedestination <path>` | Move destination (supports template variables) |
| `--force-move` | Overwrite existing files on move |
| `--force-rename` | Overwrite existing files on rename |

#### Config

| Flag | Description |
|------|-------------|
| `-c`, `--config <path>` | Load config from file |
| `-s`, `--save <path>` | Save current config to file and exit |
| `-p`, `--preview-config` | Show current config and exit |
| `--version` | Show version |

## Configuration

tvn looks for config files in this order:

1. Path specified with `--config`
2. `~/.config/tvn/tvn.json`
3. `~/.tvn.json` (legacy)

Generate a default config file:

```bash
tvn --save=~/.config/tvn/tvn.json
```

Config files from the original Python tvnamer using `%(key)s`-style placeholders in filename templates are supported and converted automatically.

### Key Configuration Options

#### Output Filename Templates

Templates use Go `{{.Field}}` syntax (or Python `%(key)s` for compatibility). Available fields: `SeriesName`, `SeasonNumber`, `Episode`, `EpisodeName`, `Ext`, `Group`, `CRC`.

| Option | Default |
|--------|---------|
| `filename_with_episode` | `{{.SeriesName}}.S{{.SeasonNumber}}E{{.Episode}}.{{.EpisodeName}}{{.Ext}}` |
| `filename_without_episode` | `{{.SeriesName}}.{{.SeasonNumber}}x{{.Episode}}{{.Ext}}` |
| `filename_with_episode_no_season` | `{{.SeriesName}}.E{{.Episode}}.{{.EpisodeName}}{{.Ext}}` |
| `filename_with_date_and_episode` | `{{.SeriesName}}.{{.Episode}}.{{.EpisodeName}}{{.Ext}}` |
| `filename_anime_with_episode` | `[{{.Group}}] {{.SeriesName}} - {{.Episode}} - {{.EpisodeName}} [{{.CRC}}]{{.Ext}}` |

#### Replacements

Replacement rules are lists of `{"is_regex": bool, "match": "...", "replacement": "..."}` objects.

| Option | Description |
|--------|-------------|
| `input_filename_replacements` | Applied to filename before parsing |
| `output_filename_replacements` | Applied to output filename after template rendering |
| `move_files_fullpath_replacements` | Applied to the full destination path when moving |
| `input_series_replacements` | Regex map applied to parsed series name before TVDB lookup |
| `output_series_replacements` | String map applied to TVDB series name in output |

Default `output_filename_replacements`:
```json
[
  {"is_regex": true, "match": "[ ]",            "replacement": "."},
  {"is_regex": true, "match": "[':?!@#$%^&*]",  "replacement": ""},
  {"is_regex": true, "match": "[&]",             "replacement": "and"}
]
```

#### Episode Formatting

| Option | Default | Description |
|--------|---------|-------------|
| `episode_single` | `"%02d"` | Format string for a single episode number |
| `episode_separator` | `"E"` | Joiner for multi-episode numbers (e.g., `01E02`) |
| `multiep_join_name_with` | `", "` | Joiner when episode names differ |
| `multiep_format` | `({{.EpisodeMin}}-{{.EpisodeMax}}).{{.EpName}}` | Format when multi-ep names are the same |

#### Behavior

| Option | Default | Description |
|--------|---------|-------------|
| `always_rename` | `false` | Auto-rename without prompting |
| `select_first` | `false` | Auto-select first TVDB result |
| `batch` | `false` | Same as `always_rename` + `select_first` |
| `dry_run` | `false` | Preview only, no changes made |
| `recursive` | `false` | Recurse into subdirectories |
| `skip_file_on_error` | `true` | Skip files that error rather than aborting |
| `skip_behaviour` | `"skip"` | `"skip"` to skip errors, `"error"` to abort |
| `overwrite_destination_on_rename` | `false` | Overwrite if target already exists |
| `overwrite_destination_on_move` | `false` | Overwrite move target if it exists |

#### File Discovery

| Option | Default | Description |
|--------|---------|-------------|
| `valid_extensions` | `[]` | Restrict to these extensions (e.g., `["mkv","mp4","avi"]`); empty = all |
| `filename_blacklist` | `[]` | Skip files matching these (literal or regex) |
| `extension_pattern` | `(\.[a-zA-Z0-9]+)$` | Regex for splitting basename from extension |

#### Filename Sanitization

| Option | Default | Description |
|--------|---------|-------------|
| `normalize_unicode_filenames` | `true` | Replace accented chars with ASCII equivalents |
| `replace_invalid_characters_with` | `""` | Replacement for invalid characters |
| `windows_safe_filenames` | `false` | Force Windows-safe output |
| `lowercase_filename` | `false` | Lowercase entire output filename |
| `titlecase_filename` | `false` | Title-case entire output filename |
| `custom_filename_character_blacklist` | `""` | Additional characters to strip/replace |

#### Move Files

| Option | Default | Description |
|--------|---------|-------------|
| `move_files_enable` | `false` | Move files to destination after rename |
| `move_files_destination` | `"."` | Destination path template. Vars: `{{.SeriesName}}`, `{{.SeasonNumber}}`, `{{.Episode}}`, `{{.OriginalFilename}}` |
| `move_files_destination_date` | `"."` | Destination for date-based episodes. Vars: `{{.SeriesName}}`, `{{.Year}}`, `{{.Month}}`, `{{.Day}}`, `{{.OriginalFilename}}` |
| `move_files_confirmation` | `true` | Prompt before moving |
| `move_files_only` | `false` | Move without renaming |
| `always_move` | `false` | Delete source after cross-device copy |
| `leave_symlink` | `false` | Leave a symlink at the original path |

#### TheTVDB / Metadata

| Option | Default | Description |
|--------|---------|-------------|
| `language` | `"en"` | Metadata language (2-letter code) |
| `search_all_languages` | `true` | Return multilingual results |
| `order` | `"aired"` | Episode order: `"aired"` or `"dvd"` |
| `series_id` | `null` | Force a specific TheTVDB series ID |
| `force_name` | `null` | Force series name (skips TVDB name lookup) |
| `tvdb_api_key` | `null` | Custom API key (default key is bundled) |

### Example Config Snippets

Move files to a `Series/Season N/` folder structure:

```json
{
  "move_files_enable": true,
  "move_files_destination": "/media/tv/{{.SeriesName}}/Season {{.SeasonNumber}}/"
}
```

Use French metadata:

```json
{
  "language": "fr"
}
```

Only process video files:

```json
{
  "valid_extensions": ["mkv", "mp4", "avi", "m4v"]
}
```

## Supported Filename Formats

| Format | Example |
|--------|---------|
| Standard SxxExx | `brass.eye.s01e01.avi` |
| Season x Episode | `foo.1x23.avi` |
| Multi-episode | `scrubs.s01e23e24.avi`, `foo.s01e23-24.avi` |
| Bracketed | `foo.[1x09-11].avi`, `foo - [012].avi` |
| Anime | `[Shinsen-Subs] Beet - 19 [24DAB497].mkv` |
| Date-based | `show.2010.01.02.etc` |
| Part notation | `show name 2 of 6` |
| Season/Episode words | `show name Season 01 Episode 20` |

## Development

```bash
make build        # Build the binary
make install      # Build and install via go install
make test         # Run tests
make test-verbose # Run tests with verbose output
make lint         # Run fmt + vet
make fmt          # Format code
make vet          # Run go vet
make clean        # Remove build artifacts
make push         # Bump patch version, commit, tag, and push
```

Set `TVN_TEST_MODE=1` to use cached test fixtures instead of live TVDB API calls (useful for CI and offline development).

## License

[Unlicense](https://unlicense.org)
