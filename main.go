// Command opendoc is a terminal UI for opendoc — a CLI-based open-source
// documentation manager. It loads the user config and hands control to the
// Bubble Tea program, which routes to the setup wizard or the main menu
// depending on whether setup has been completed.
package main

import (
	"fmt"
	"os"

	"opendoc/config"
	"opendoc/ui"

	tea "charm.land/bubbletea/v2"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(ui.NewSplashModel(cfg))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error running program: %v\n", err)
		os.Exit(1)
	}
}
