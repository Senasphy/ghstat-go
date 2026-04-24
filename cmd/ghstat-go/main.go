package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/senasphy/ghstat-go/internal/githubapi"
	"github.com/senasphy/ghstat-go/internal/ui"
)

type options struct {
	Year  int
	Token string
}

var errNoDefaultLogin = errors.New("no GitHub username argument and no local GitHub login found")

func main() {
	login, opts, err := parseOptions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghstat: %v\n", err)
		if !errors.Is(err, errNoDefaultLogin) {
			fmt.Fprintln(os.Stderr)
			flag.Usage()
		}
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

	if flag.NArg() > 1 {
		return "", opts, errors.New("at most one GitHub username is allowed")
	}

	var login string
	if flag.NArg() == 1 {
		login = strings.TrimSpace(flag.Arg(0))
		if login == "" {
			return "", opts, errors.New("GitHub username cannot be empty")
		}
	} else {
		resolved, ok := resolveDefaultLogin()
		if !ok {
			return "", opts, errNoDefaultLogin
		}
		login = resolved
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

func resolveDefaultLogin() (string, bool) {
	envCandidates := []string{
		"GITHUB_USER",
		"GH_USER",
		"GITHUB_USERNAME",
	}
	for _, key := range envCandidates {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value, true
		}
	}

	if value, ok := runAndTrim("git", "config", "--get", "github.user"); ok {
		return value, true
	}
	if value, ok := runAndTrim("git", "config", "--global", "--get", "github.user"); ok {
		return value, true
	}

	if value, ok := runAndTrim("gh", "api", "user", "-q", ".login"); ok {
		return value, true
	}

	return "", false
}

func runAndTrim(name string, args ...string) (string, bool) {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return "", false
	}
	value := strings.TrimSpace(string(out))
	if value == "" {
		return "", false
	}
	return value, true
}
