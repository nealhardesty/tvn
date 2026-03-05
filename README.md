# tvn

A command-line utility that automatically renames TV episode files by fetching metadata from [TheTVDB.com](https://thetvdb.com). Transforms messy download filenames like `some.show.s01e03.blah.abc.avi` into clean, consistent names like `Some Show - [01x03] - The Episode Name.avi`.

Go reimplementation of the Python [tvnamer](https://github.com/dbr/tvnamer) project.

**Note:** TV information is provided by TheTVDB.com. tvn is not endorsed or certified by TheTVDB.com or its affiliates.

## Installation

Requires Go 1.22+.

```bash
go install tvn@latest
```

Or build from source:

```bash
git clone https://github.com/your-username/tvn.git
cd tvn
make build
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

# Move files to organized directory structure
tvn -m -d "/media/tv/{{.SeriesName}}/Season {{.SeasonNumber}}/" *.mkv
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
| `-f`, `--selectfirst` | Auto-select first search result |
| `-b`, `--batch` | No prompts (`--always` + `--selectfirst`) |

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
| `-m`, `--move` | Move files to destination after rename |
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

See [PRD.md](PRD.md) for the full configuration reference.

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
make build       # Build the binary
make test        # Run tests
make lint        # Run fmt + vet
make clean       # Remove build artifacts
make install     # Install via go install
```

## License

[Unlicense](https://unlicense.org)
