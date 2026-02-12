package main

import (
	"context"
	"fmt"
	"os"

	"terminal-business/internal/persistence"
	"terminal-business/internal/platform/clock"
	"terminal-business/internal/platform/random"
	"terminal-business/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) != 1 {
		fmt.Fprintln(os.Stderr, "tbiz accepts no flags or subcommands")
		os.Exit(2)
	}
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() (err error) {
	defer func() {
		if r := recover(); r != nil {
			panicModel := ui.NewPanicModel(fmt.Sprintf("Recovered panic during startup: %v", r))
			_, _ = tea.NewProgram(panicModel, tea.WithAltScreen()).Run()
			err = fmt.Errorf("panic recovered: %v", r)
		}
	}()

	baseDir, err := persistence.DefaultBaseDir()
	if err != nil {
		return err
	}
	store := persistence.NewLocalStore(baseDir)
	entries, err := store.LoadIndex(context.Background())
	if err != nil {
		entries = []persistence.SaveIndexEntry{}
	}

	model := ui.NewModel(store, clock.System{}, random.New(), entries)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
