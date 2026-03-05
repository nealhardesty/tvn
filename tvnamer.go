package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// ProcessFiles is the main entry point for processing TV episode files.
func ProcessFiles(cfg *Config, files []string, tvdb *TVDBClient) error {
	parser, err := NewParser(cfg)
	if err != nil {
		return fmt.Errorf("initializing parser: %w", err)
	}

	// Resolve files from paths (may include directories)
	allFiles, err := discoverFiles(cfg, files)
	if err != nil {
		return err
	}

	if len(allFiles) == 0 {
		fmt.Println("No valid files found.")
		return nil
	}

	// Cache for series lookups
	type seriesInfo struct {
		series   *SeriesSearchResult
		episodes []Episode
	}
	seriesCache := make(map[string]*seriesInfo)

	alwaysRename := cfg.AlwaysRename || cfg.Batch

	for _, filePath := range allFiles {
		filename := filepath.Base(filePath)
		dir := filepath.Dir(filePath)

		// Apply input filename replacements
		processedName := applyReplacements(filename, cfg.InputFilenameReplacements)

		// Parse filename
		parsed, err := parser.Parse(processedName)
		if err != nil {
			if cfg.Verbose {
				fmt.Printf("# Skipping (no match): %s\n", filename)
			}
			if cfg.SkipBehaviour == "error" {
				return fmt.Errorf("could not parse: %s", filename)
			}
			continue
		}

		if cfg.Verbose {
			fmt.Printf("# Parsed: %s -> %+v\n", filename, parsed)
		}

		// Apply input series replacements
		seriesName := parsed.SeriesName
		for pattern, replacement := range cfg.InputSeriesReplacements {
			re, err := regexp.Compile(pattern)
			if err != nil {
				continue
			}
			seriesName = re.ReplaceAllString(seriesName, replacement)
		}
		parsed.SeriesName = seriesName

		// Determine series name for lookup
		lookupName := seriesName
		if cfg.ForceName != nil {
			lookupName = *cfg.ForceName
		}

		// Get series info (cached)
		cacheKey := strings.ToLower(lookupName)
		info, cached := seriesCache[cacheKey]
		if !cached {
			info = &seriesInfo{}

			if cfg.SeriesID != nil {
				// Use provided series ID
				series, err := tvdb.GetSeriesByID(*cfg.SeriesID)
				if err != nil {
					return handleError(cfg, filename, fmt.Errorf("looking up series ID %d: %w", *cfg.SeriesID, err))
				}
				info.series = series
			} else if cfg.ForceName != nil {
				// Use forced name without TVDB lookup for the name
				// Still need to search to get an ID for episode data
				results, err := tvdb.SearchSeries(*cfg.ForceName)
				if err != nil {
					return handleError(cfg, filename, fmt.Errorf("searching TVDB: %w", err))
				}
				if len(results) > 0 {
					info.series = &results[0]
					info.series.Name = *cfg.ForceName
				}
			} else {
				results, err := tvdb.SearchSeries(lookupName)
				if err != nil {
					if skipErr := handleError(cfg, filename, fmt.Errorf("searching TVDB: %w", err)); skipErr != nil {
						return skipErr
					}
					continue
				}
				if len(results) == 0 {
					if skipErr := handleError(cfg, filename, fmt.Errorf("no TVDB results for: %s", lookupName)); skipErr != nil {
						return skipErr
					}
					continue
				}

				if len(results) == 1 || cfg.SelectFirst || cfg.Batch {
					info.series = &results[0]
				} else {
					selected, err := PromptSeriesSelection(results)
					if err != nil {
						return err
					}
					if selected == nil {
						fmt.Println("Quitting.")
						return nil
					}
					info.series = selected
				}
			}

			if info.series != nil {
				// Fetch episodes
				episodes, err := tvdb.GetEpisodes(info.series.TVDBID, cfg.Order)
				if err != nil {
					if skipErr := handleError(cfg, filename, fmt.Errorf("fetching episodes: %w", err)); skipErr != nil {
						return skipErr
					}
					continue
				}
				info.episodes = episodes
			}

			seriesCache[cacheKey] = info
		}

		if info.series == nil {
			continue
		}

		// Apply output series replacements
		outputSeriesName := info.series.Name
		for match, replacement := range cfg.OutputSeriesReplacements {
			outputSeriesName = strings.ReplaceAll(outputSeriesName, match, replacement)
		}

		// Look up episode name(s)
		var episodeNames []string
		if parsed.EpisodeType == EpisodeDateBased {
			ep := FindEpisodeByDate(info.episodes, parsed.Year, parsed.Month, parsed.Day)
			if ep != nil {
				episodeNames = []string{ep.Name}
				// Set episode numbers from the found episode for formatting
				parsed.EpisodeNumbers = []int{ep.Number}
				parsed.SeasonNumber = ep.SeasonNumber
				parsed.HasSeason = true
			}
		} else {
			for _, epNum := range parsed.EpisodeNumbers {
				ep := FindEpisode(info.episodes, parsed.SeasonNumber, epNum)
				if ep != nil {
					episodeNames = append(episodeNames, ep.Name)
				} else {
					episodeNames = append(episodeNames, "")
				}
			}
		}

		// Format output filename
		newFilename, err := FormatFilename(cfg, parsed, outputSeriesName, episodeNames)
		if err != nil {
			if skipErr := handleError(cfg, filename, fmt.Errorf("formatting filename: %w", err)); skipErr != nil {
				return skipErr
			}
			continue
		}

		// Apply output filename replacements
		newFilename = applyReplacements(newFilename, cfg.OutputFilenameReplacements)

		// Sanitize
		newFilename = SanitizeFilename(newFilename, cfg)

		// Check if rename is needed
		if newFilename == filename {
			if cfg.Verbose {
				fmt.Printf("# Already named correctly: %s\n", filename)
			}
			continue
		}

		newPath := filepath.Join(dir, newFilename)

		if cfg.DryRun {
			fmt.Printf("%s -> %s\n", filename, newFilename)
			continue
		}

		// Prompt for rename
		if !alwaysRename {
			action := PromptRename(filename, newFilename)
			switch action {
			case RenameNo:
				continue
			case RenameAlways:
				alwaysRename = true
			case RenameQuit:
				return nil
			case RenameYes:
				// proceed
			}
		} else {
			fmt.Printf("%s -> %s\n", filename, newFilename)
		}

		// Perform rename
		if !cfg.MoveFilesOnly {
			if err := RenameFile(filePath, newPath, cfg.OverwriteDestinationOnRename); err != nil {
				if skipErr := handleError(cfg, filename, fmt.Errorf("renaming: %w", err)); skipErr != nil {
					return skipErr
				}
				continue
			}
			filePath = newPath // update for potential move
		}

		// Move file if enabled
		if cfg.MoveFilesEnable {
			moveDest, err := computeMoveDest(cfg, parsed, outputSeriesName, filePath)
			if err != nil {
				if skipErr := handleError(cfg, filename, fmt.Errorf("computing move destination: %w", err)); skipErr != nil {
					return skipErr
				}
				continue
			}

			moveDest = applyReplacements(moveDest, cfg.MoveFilesFullpathReplacements)

			if cfg.MoveFilesLowercaseDestination {
				// Lowercase just the directory parts, not the filename
				moveDir := filepath.Dir(moveDest)
				moveBase := filepath.Base(moveDest)
				moveDest = filepath.Join(strings.ToLower(moveDir), moveBase)
			}

			if cfg.DryRun {
				fmt.Printf("  -> move to: %s\n", moveDest)
				continue
			}

			doMove := true
			if cfg.MoveFilesConfirmation && !alwaysRename {
				action := PromptMove(filePath, moveDest)
				switch action {
				case MoveNo:
					doMove = false
				case MoveQuit:
					return nil
				}
			}

			if doMove {
				if err := MoveFile(filePath, moveDest, cfg.OverwriteDestinationOnMove, cfg.AlwaysMove, cfg.LeaveSymlink); err != nil {
					if skipErr := handleError(cfg, filename, fmt.Errorf("moving: %w", err)); skipErr != nil {
						return skipErr
					}
				}
			}
		}
	}

	return nil
}

// discoverFiles resolves file paths from arguments (files and directories).
func discoverFiles(cfg *Config, paths []string) ([]string, error) {
	var files []string

	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, fmt.Errorf("stat %s: %w", p, err)
		}

		if !info.IsDir() {
			if isValidFile(cfg, info.Name()) {
				abs, _ := filepath.Abs(p)
				files = append(files, abs)
			}
			continue
		}

		// Directory: walk or list
		if cfg.Recursive {
			err = filepath.WalkDir(p, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}
				if isValidFile(cfg, d.Name()) {
					abs, _ := filepath.Abs(path)
					files = append(files, abs)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			entries, err := os.ReadDir(p)
			if err != nil {
				return nil, fmt.Errorf("reading directory %s: %w", p, err)
			}
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				if isValidFile(cfg, e.Name()) {
					abs, _ := filepath.Abs(filepath.Join(p, e.Name()))
					files = append(files, abs)
				}
			}
		}
	}

	return files, nil
}

// isValidFile checks if a file should be processed based on config.
func isValidFile(cfg *Config, filename string) bool {
	// Check extension filter
	if len(cfg.ValidExtensions) > 0 {
		ext := strings.TrimPrefix(filepath.Ext(filename), ".")
		ext = strings.ToLower(ext)
		found := false
		for _, valid := range cfg.ValidExtensions {
			if strings.ToLower(valid) == ext {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check blacklist
	for _, bl := range cfg.FilenameBlacklist {
		if bl == filename {
			return false
		}
		// Try as regex
		if re, err := regexp.Compile(bl); err == nil {
			if re.MatchString(filename) {
				return false
			}
		}
	}

	return true
}

// computeMoveDest computes the move destination path.
func computeMoveDest(cfg *Config, parsed *ParseResult, seriesName, currentPath string) (string, error) {
	tmplStr := cfg.MoveFilesDestination
	if parsed.EpisodeType == EpisodeDateBased {
		tmplStr = cfg.MoveFilesDestinationDate
	}

	tmpl, err := template.New("movepath").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("parsing move destination template: %w", err)
	}

	data := EpisodeData{
		SeriesName:       seriesName,
		SeasonNumber:     fmt.Sprintf("%d", parsed.SeasonNumber),
		Episode:          formatEpisodeNumbers(cfg, parsed),
		OriginalFilename: filepath.Base(currentPath),
	}
	if parsed.EpisodeType == EpisodeDateBased {
		data.Year = fmt.Sprintf("%04d", parsed.Year)
		data.Month = fmt.Sprintf("%02d", parsed.Month)
		data.Day = fmt.Sprintf("%02d", parsed.Day)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	dest := buf.String()
	if cfg.MoveFilesDestinationIsFilepath {
		return dest, nil
	}
	return filepath.Join(dest, filepath.Base(currentPath)), nil
}

// handleError handles an error according to skip settings.
// Returns nil if the file should be skipped, or the error if processing should stop.
func handleError(cfg *Config, filename string, err error) error {
	fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", filename, err)
	if cfg.SkipFileOnError && cfg.SkipBehaviour == "skip" {
		return nil
	}
	return err
}
