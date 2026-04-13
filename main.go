package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"

	tea "github.com/charmbracelet/bubbletea"
)

var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	demo := flag.Bool("demo", false, "run with fake sessions for screenshots (no DB required)")

	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage: lazyopencode [flags]\n\n")                                //nolint:errcheck // writing usage to stderr; failure is not actionable
		fmt.Fprintf(w, "A terminal UI for browsing and managing opencode sessions.\n\n") //nolint:errcheck // writing usage to stderr; failure is not actionable
		fmt.Fprintf(w, "Flags:\n")                                                       //nolint:errcheck // writing usage to stderr; failure is not actionable
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintf(tw, "  --help\tshow this help message\n") //nolint:errcheck // writing usage to stderr; failure is not actionable
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(tw, "  --%s\t%s\n", f.Name, f.Usage) //nolint:errcheck // writing usage to stderr; failure is not actionable
		})
		tw.Flush() //nolint:errcheck // flushing usage output to stderr; failure is not actionable
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("lazyopencode %s\n", version)
		os.Exit(0)
	}

	out, err := exec.Command("opencode", "db", "path").Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "lazyopencode: could not resolve opencode DB path (is opencode in PATH?): %v\n", err)
		os.Exit(1)
	}
	dbPath := strings.TrimSpace(string(out))
	if dbPath == "" {
		fmt.Fprintf(os.Stderr, "lazyopencode: opencode db path returned empty output\n")
		os.Exit(1)
	}

	p := tea.NewProgram(
		newModel(dbPath, *demo),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "lazyopencode: %v\n", err)
		os.Exit(1)
	}
}
