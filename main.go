package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

const Version = "1.0.2"

func main() {
	var showVersion = flag.Bool("version", false, "Show version information")
	var showHelp = flag.Bool("help", false, "Show help information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("OCLI v%s - Terminal Outliner\n", Version)
		fmt.Println("Built with ❤️  using Go and Charm TUI libraries")
		return
	}

	if *showHelp {
		fmt.Println("OCLI - Terminal Outliner")
		fmt.Printf("Version: %s\n\n", Version)
		fmt.Println("Usage: ocli [options]")
		fmt.Println("\nOptions:")
		fmt.Println("  --version    Show version information")
		fmt.Println("  --help       Show this help message")
		fmt.Println("\nKeyboard shortcuts available in the app:")
		fmt.Println("  h            Show interactive help screen")
		fmt.Println("  s            Show settings")
		fmt.Println("  q            Quit application")
		fmt.Println("\nData is automatically saved to ~/.config/ocli/data.json")
		return
	}

	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
