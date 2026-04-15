package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"ghstat-go/internal/githubapi"
	"ghstat-go/internal/ui"
)

type options struct {
	Year  int
	Token string
}

func main() {
	login, opts, err := parseOptions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghstat: %v\n\n", err)
		flag.Usage()
		os.Exit(2)
	}

	service, err := githubapi.NewService(opts.Token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghstat: %v\n", err)
		os.Exit(1)
	}

	program := tea.NewProgram(
		ui.NewModel(service, login, opts.Year),
		tea.WithAltScreen(),
	)

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "ghstat: %v\n", err)
		os.Exit(1)
	}
}

func parseOptions() (string, options, error) {
	var opts options

	flag.IntVar(&opts.Year, "year", 0, "end year for the rolling 12-month window (defaults to the current year)")
	flag.StringVar(&opts.Token, "token", "", "GitHub token (defaults to GITHUB_TOKEN)")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [flags] <github-username>\n\n", os.Args[0])
		fmt.Fprintln(flag.CommandLine.Output(), "Interactive GitHub contribution charts for the terminal.")
		fmt.Fprintln(flag.CommandLine.Output())
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() != 1 {
		return "", opts, errors.New("exactly one GitHub username is required")
	}

	login := strings.TrimSpace(flag.Arg(0))
	if login == "" {
		return "", opts, errors.New("GitHub username cannot be empty")
	}

	if opts.Year < 0 {
		return "", opts, errors.New("year must be positive")
	}

	currentYear := time.Now().Year()
	if opts.Year > currentYear {
		return "", opts, fmt.Errorf("year %d is in the future", opts.Year)
	}

	return login, opts, nil
}
