package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var stdinReader = bufio.NewReader(os.Stdin)

// PromptSeriesSelection asks the user to pick a series from results.
// Returns the selected series, or nil if user quits.
func PromptSeriesSelection(results []SeriesSearchResult) (*SeriesSearchResult, error) {
	fmt.Println("TVDB Search Results:")
	for i, r := range results {
		lang := r.PrimaryLanguage
		if lang == "" {
			lang = "??"
		}
		year := r.Year
		if year == "" {
			year = "????"
		}
		fmt.Printf(" %d -> %s [%s] (%s)\n", i+1, r.Name, lang, year)
	}
	for {
		fmt.Printf("Enter choice (1-%d, q to quit): ", len(results))
		line, err := stdinReader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "q" || line == "Q" {
			return nil, nil
		}
		choice, err := strconv.Atoi(line)
		if err != nil || choice < 1 || choice > len(results) {
			fmt.Println("Invalid choice, try again.")
			continue
		}
		return &results[choice-1], nil
	}
}

// RenameAction represents the user's choice for a rename prompt.
type RenameAction int

const (
	RenameYes RenameAction = iota
	RenameNo
	RenameAlways
	RenameQuit
)

// PromptRename asks the user to confirm a rename operation.
func PromptRename(oldName, newName string) RenameAction {
	fmt.Printf("Old filename: %s\n", oldName)
	fmt.Printf("New filename: %s\n", newName)
	for {
		fmt.Print("Rename? (y/n/a/q): ")
		line, err := stdinReader.ReadString('\n')
		if err != nil {
			return RenameQuit
		}
		switch strings.TrimSpace(strings.ToLower(line)) {
		case "y", "yes":
			return RenameYes
		case "n", "no":
			return RenameNo
		case "a", "always":
			return RenameAlways
		case "q", "quit":
			return RenameQuit
		default:
			fmt.Println("Invalid choice. Enter y, n, a, or q.")
		}
	}
}

// MoveAction represents the user's choice for a move prompt.
type MoveAction int

const (
	MoveYes MoveAction = iota
	MoveNo
	MoveQuit
)

// PromptMove asks the user to confirm a move operation.
func PromptMove(src, dst string) MoveAction {
	fmt.Printf("Move to: %s\n", dst)
	for {
		fmt.Print("Move? (y/n/q): ")
		line, err := stdinReader.ReadString('\n')
		if err != nil {
			return MoveQuit
		}
		switch strings.TrimSpace(strings.ToLower(line)) {
		case "y", "yes":
			return MoveYes
		case "n", "no":
			return MoveNo
		case "q", "quit":
			return MoveQuit
		default:
			fmt.Println("Invalid choice. Enter y, n, or q.")
		}
	}
}
