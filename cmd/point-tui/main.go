package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"github.com/dariy/point-tui/internal/api"
	"github.com/dariy/point-tui/internal/config"
	"github.com/dariy/point-tui/internal/ui"
)

func main() {
	var login bool
	flag.BoolVar(&login, "login", false, "Prompt for credentials to access private content")
	flag.BoolVar(&login, "l", false, "Prompt for credentials to access private content")
	flag.Parse()

	baseURL := ""
	if flag.NArg() > 0 {
		baseURL = flag.Arg(0)
	}

	cfg, err := config.Load(baseURL, login)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	app := ui.NewApp(cfg.BaseURL, cfg.Login)

	if cfg.Login {
		client, err := api.NewClient(cfg.BaseURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating client: %v\n", err)
			os.Exit(1)
		}
		if err := promptLogin(client); err != nil {
			fmt.Fprintf(os.Stderr, "login failed: %v\n", err)
			os.Exit(1)
		}
		app.SetClient(client)
	}

	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error running program: %v\n", err)
		os.Exit(1)
	}
}

func promptLogin(client *api.Client) error {
	fmt.Fprint(os.Stderr, "Password: ")
	passBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr) // newline after hidden input
	if err != nil {
		return fmt.Errorf("reading password: %w", err)
	}

	return client.Login(context.Background(), "", string(passBytes), true)
}
