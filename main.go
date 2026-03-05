package main

import (
	"encoding/json"
	"fmt"
	"os"

	pflag "github.com/spf13/pflag"
)

var version = "dev"

func main() {
	// CLI flags
	var (
		verbose        bool
		notVerbose     bool
		dryRun         bool
		always         bool
		notAlways      bool
		selectFirst    bool
		notSelectFirst bool
		batch          bool
		notBatch       bool
		configPath     string
		savePath       string
		previewConfig  bool
		nameOverride   string
		seriesID       int
		order          string
		lang           string
		recursive      bool
		notRecursive   bool
		moveFiles      bool
		notMove        bool
		forceMove      bool
		forceRename    bool
		moveDest       string
		showVersion    bool
	)

	pflag.BoolVarP(&verbose, "verbose", "v", false, "Show debugging info")
	pflag.BoolVarP(&notVerbose, "not-verbose", "q", false, "Disable verbose")
	pflag.BoolVar(&dryRun, "dry-run", false, "Preview changes without renaming")
	pflag.BoolVarP(&always, "always", "a", false, "Always rename after series selection")
	pflag.BoolVar(&notAlways, "not-always", false, "Override --always")
	pflag.BoolVarP(&selectFirst, "selectfirst", "f", false, "Auto-select first search result")
	pflag.BoolVar(&notSelectFirst, "not-selectfirst", false, "Override --selectfirst")
	pflag.BoolVarP(&batch, "batch", "b", false, "No prompts (--always + --selectfirst)")
	pflag.BoolVar(&notBatch, "not-batch", false, "Override --batch")
	pflag.StringVarP(&configPath, "config", "c", "", "Config file path")
	pflag.StringVarP(&savePath, "save", "s", "", "Save config to file and exit")
	pflag.BoolVarP(&previewConfig, "preview-config", "p", false, "Show config and exit")
	pflag.StringVarP(&nameOverride, "name", "n", "", "Override series name")
	pflag.IntVar(&seriesID, "series-id", 0, "Force TheTVDB series ID")
	pflag.StringVar(&order, "order", "", "Episode order: aired or dvd")
	pflag.StringVarP(&lang, "lang", "l", "", "Metadata language (e.g., en, de, fr)")
	pflag.BoolVarP(&recursive, "recursive", "r", false, "Recurse into subdirectories")
	pflag.BoolVar(&notRecursive, "not-recursive", false, "Disable recursive")
	pflag.BoolVarP(&moveFiles, "move", "m", false, "Move files to destination")
	pflag.BoolVar(&notMove, "not-move", false, "Disable move")
	pflag.BoolVar(&forceMove, "force-move", false, "Overwrite on move")
	pflag.BoolVar(&forceRename, "force-rename", false, "Overwrite on rename")
	pflag.StringVarP(&moveDest, "movedestination", "d", "", "Move destination path")
	pflag.BoolVar(&showVersion, "version", false, "Show version")

	pflag.Parse()

	if showVersion {
		fmt.Printf("tvnamer %s\n", version)
		os.Exit(0)
	}

	// Load config
	cfg := DefaultConfig()
	if configPath != "" {
		loaded, err := LoadConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
		cfg = loaded
	} else if found := FindConfigFile(); found != "" {
		loaded, err := LoadConfig(found)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: error loading config %s: %v\n", found, err)
		} else {
			cfg = loaded
		}
	}

	// Apply CLI overrides
	if pflag.CommandLine.Changed("verbose") {
		cfg.Verbose = verbose
	}
	if notVerbose {
		cfg.Verbose = false
	}
	if pflag.CommandLine.Changed("dry-run") {
		cfg.DryRun = dryRun
	}
	if pflag.CommandLine.Changed("always") {
		cfg.AlwaysRename = always
	}
	if notAlways {
		cfg.AlwaysRename = false
	}
	if pflag.CommandLine.Changed("selectfirst") {
		cfg.SelectFirst = selectFirst
	}
	if notSelectFirst {
		cfg.SelectFirst = false
	}
	if pflag.CommandLine.Changed("batch") {
		cfg.Batch = batch
		if batch {
			cfg.AlwaysRename = true
			cfg.SelectFirst = true
		}
	}
	if notBatch {
		cfg.Batch = false
	}
	if pflag.CommandLine.Changed("recursive") {
		cfg.Recursive = recursive
	}
	if notRecursive {
		cfg.Recursive = false
	}
	if pflag.CommandLine.Changed("move") {
		cfg.MoveFilesEnable = moveFiles
	}
	if notMove {
		cfg.MoveFilesEnable = false
	}
	if forceMove {
		cfg.OverwriteDestinationOnMove = true
	}
	if forceRename {
		cfg.OverwriteDestinationOnRename = true
	}
	if nameOverride != "" {
		cfg.ForceName = &nameOverride
	}
	if seriesID != 0 {
		cfg.SeriesID = &seriesID
	}
	if order != "" {
		cfg.Order = order
	}
	if lang != "" {
		cfg.Language = lang
	}
	if moveDest != "" {
		cfg.MoveFilesDestination = moveDest
		cfg.MoveFilesEnable = true
	}

	// Save config and exit
	if savePath != "" {
		if err := SaveConfig(cfg, savePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Config saved to %s\n", savePath)
		os.Exit(0)
	}

	// Preview config and exit
	if previewConfig {
		data, _ := json.MarshalIndent(cfg, "", "  ")
		fmt.Println(string(data))
		os.Exit(0)
	}

	// Remaining args are files/directories
	args := pflag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: tvnamer [options] <files or directories>")
		pflag.PrintDefaults()
		os.Exit(1)
	}

	// Initialize TVDB client
	tvdb := NewTVDBClient(&cfg)
	if err := tvdb.Login(); err != nil {
		fmt.Fprintf(os.Stderr, "TVDB error: %v\n", err)
		os.Exit(1)
	}

	// Process files
	if err := ProcessFiles(&cfg, args, tvdb); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
