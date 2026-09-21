// Package main is the entry point for vo-studio-tui.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	projectFlag := flag.String("project", "", "Path to voiceover project directory")
	flag.Parse()

	targetDir := *projectFlag
	if targetDir == "" && flag.NArg() > 0 {
		targetDir = flag.Arg(0)
	}

	regPath := filepath.Join(DefaultDataDir(), "registry.json")
	reg, err := LoadRegistry(regPath)
	if err != nil {
		reg = &Registry{}
	}

	configPath := filepath.Join(DefaultConfigDir(), "config.json")
	globalCfg, err := LoadGlobalConfig(configPath)
	if err != nil {
		globalCfg = DefaultGlobalConfig()
	}

	var initialModel *AppModel

	if targetDir != "" {
		absPath, err := filepath.Abs(targetDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to resolve project path: %v\n", err)
			os.Exit(1)
		}

		proj, err := LoadProject(absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load project from %s: %v\n", absPath, err)
			os.Exit(1)
		}
		proj.ApplyGlobalConfig(globalCfg)

		reg.Register(proj.Name, proj.RootPath)
		_ = SaveRegistry(regPath, reg)

		initialModel = NewAppModel(proj)
		initialModel.Registry = reg
		initialModel.RegistryPath = regPath
		initialModel.GlobalConfig = globalCfg
	} else {
		initialModel = NewHubModel(reg, regPath)
		initialModel.GlobalConfig = globalCfg
	}


	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Application error: %v\n", err)
		os.Exit(1)
	}
}
