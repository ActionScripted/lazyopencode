package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	tea "github.com/charmbracelet/bubbletea"
)

var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	demo := flag.Bool("demo", false, "run with fake sessions for screenshots (no DB required)")

	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage: lazyopencode [flags]\n\n")                                //nolint:errcheck
		fmt.Fprintf(w, "A terminal UI for browsing and managing opencode sessions.\n\n") //nolint:errcheck
		fmt.Fprintf(w, "Flags:\n")                                                       //nolint:errcheck
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintf(tw, "  --help\tshow this help message\n") //nolint:errcheck
		flag.VisitAll(func(f *flag.Flag) {                    //nolint:errcheck
			fmt.Fprintf(tw, "  --%s\t%s\n", f.Name, f.Usage) //nolint:errcheck
		})
		tw.Flush() //nolint:errcheck
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("lazyopencode %s\n", version)
		os.Exit(0)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "lazyopencode: could not resolve home directory: %v\n", err)
		os.Exit(1)
	}
	dbPath := filepath.Join(home, ".local", "share", "opencode", "opencode.db")

	p := tea.NewProgram(
		newModel(dbPath, *demo),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "lazyopencode: %v\n", err)
		os.Exit(1)
	}
}
