package main

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

// Terminal version and beta information
const (
	TerminalVersion = "0.9"
	IsBeta          = true
	BuildInfo       = "beta 1"
)

type Config struct {
	Prompt          string   `json:"prompt"`
	InitialCommands []string `json:"initialCommands"`
	Theme           string   `json:"theme"`
	WgetEnabled     bool     `json:"wgetEnabled"`
}

type Theme struct {
	TextColor       string `json:"textColor"`
	BackgroundColor string `json:"backgroundColor"`
	PromptColor     string `json:"promptColor"`
	ErrorColor      string `json:"errorColor"`
	OutputColor     string `json:"outputColor"`
}

var currentDir string = "."

func main() {
	config, theme := loadConfig()

	// Print welcome message and set terminal title
	getColor(theme.TextColor).Printf("\033]0;Star Shell v%s\007", TerminalVersion)
	if IsBeta {
		getColor(theme.TextColor).Printf("Welcome to Star Shell v%s %s\n\n", TerminalVersion, BuildInfo)
	} else {
		getColor(theme.TextColor).Printf("Welcome to Star Shell v%s\n\n", TerminalVersion)
	}

	// Execute initial commands
	for _, command := range config.InitialCommands {
		executeCommand(command, config, theme)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		getColor(theme.PromptColor).Printf("%s ", config.Prompt)
		scanner.Scan()
		line := scanner.Text()
		executeCommand(line, config, theme)
	}
}

func executeCommand(input string, config Config, theme Theme) {
	args := strings.Split(input, " ")

	switch strings.ToLower(args[0]) {
	case "exit":
		os.Exit(0)
	case "wget":
		if !config.WgetEnabled {
			getColor(theme.ErrorColor).Printf("wget command is disabled\n")
			return
		}
		if len(args) != 2 {
			getColor(theme.ErrorColor).Printf("wget command requires a URL\n")
			return
		}
		wget(args[1])
	case "ls":
		ls()
	case "cd":
		if len(args) != 2 {
			getColor(theme.ErrorColor).Printf("cd command requires a directory\n")
			return
		}
		cd(args[1], theme)
	case "help":
		help()
	case "verfetch":
		verfetch()
	case "ip":
		printMainIP()
	case "pkg":
		if len(args) < 3 {
			getColor(theme.ErrorColor).Printf("pkg command requires at least two arguments: install user/repo\n")
			return
		}
		if strings.ToLower(args[1]) != "install" {
			getColor(theme.ErrorColor).Printf("Unknown pkg command: %s\n", args[1])
			return
		}
		parts := strings.Split(args[2], "/")
		if len(parts) != 2 {
			getColor(theme.ErrorColor).Printf("Invalid repository format. It should be user/repo\n")
			return
		}
		pm := NewPackageManager()
		err := pm.Install(parts[0], parts[1])
		if err != nil {
			getColor(theme.ErrorColor).Printf("Failed to install: %v\n", err)
			return
		}
		getColor(theme.OutputColor).Printf("Package installation complete\n")
	default:
		cmd := exec.Command("cmd", "/C", input)
		cmd.Dir = currentDir // Set the working directory
		output, err := cmd.CombinedOutput()
		if err != nil {
			getColor(theme.ErrorColor).Printf("Error: Invalid command! %v\n", err)
			return
		}
		getColor(theme.OutputColor).Printf("%s", string(output))
	}
}

func getColor(colorName string) *color.Color {
	switch strings.ToLower(colorName) {
	case "red":
		return color.New(color.FgRed)
	case "green":
		return color.New(color.FgGreen)
	case "yellow":
		return color.New(color.FgYellow)
	case "blue":
		return color.New(color.FgBlue)
	case "magenta":
		return color.New(color.FgMagenta)
	case "cyan":
		return color.New(color.FgCyan)
	case "black":
		return color.New(color.FgBlack)
	case "white":
		return color.New(color.FgWhite)
	default:
		return color.New(color.FgWhite)
	}
}

func createDefaultConfig() Config {
	config := Config{
		Prompt:          "$",
		InitialCommands: []string{},
		Theme:           "light",
		WgetEnabled:     true,
	}

	configJson, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		color.Red("Failed to create default config: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile("config.json", configJson, 0644)
	if err != nil {
		color.Red("Failed to write default config file: %v\n", err)
		os.Exit(1)
	}

	return config
}

func createDefaultThemes() {
	themes := map[string]Theme{
		"light": {
			TextColor:       "black",
			BackgroundColor: "white",
			PromptColor:     "blue",
			ErrorColor:      "red",
			OutputColor:     "green",
		},
		"dark": {
			TextColor:       "white",
			BackgroundColor: "black",
			PromptColor:     "cyan",
			ErrorColor:      "red",
			OutputColor:     "green",
		},
	}

	for themeName, theme := range themes {
		themeJson, err := json.MarshalIndent(theme, "", "  ")
		if err != nil {
			color.Red("Failed to create %s theme: %v\n", themeName, err)
			os.Exit(1)
		}

		err = os.WriteFile("themes/"+themeName+".json", themeJson, 0644)
		if err != nil {
			color.Red("Failed to write %s theme file: %v\n", themeName, err)
			os.Exit(1)
		}
	}
}

func loadConfig() (Config, Theme) {
	var config Config

	file, err := os.ReadFile("config.json")
	if err != nil {
		// If the config file does not exist, create a default one
		if os.IsNotExist(err) {
			config = createDefaultConfig()
		} else {
			color.Red("Failed to read config file: %v\n", err)
			os.Exit(1)
		}
	} else {
		err = json.Unmarshal(file, &config)
		if err != nil {
			color.Red("Failed to parse config file: %v\n", err)
			os.Exit(1)
		}
	}

	theme := loadTheme(config.Theme)

	return config, theme
}

func loadTheme(themeName string) Theme {
	var theme Theme

	file, err := os.ReadFile("themes/" + themeName + ".json")
	if err != nil {
		// If the theme file does not exist, create default ones
		if os.IsNotExist(err) {
			// Create directory if not exist
			if _, err := os.Stat("themes"); os.IsNotExist(err) {
				os.Mkdir("themes", 0755)
			}
			// Create default themes
			createDefaultThemes()
			return loadTheme(themeName)
		} else {
			color.Red("Failed to read theme file: %v\n", err)
			os.Exit(1)
		}
	} else {
		err = json.Unmarshal(file, &theme)
		if err != nil {
			color.Red("Failed to parse theme file: %v\n", err)
			os.Exit(1)
		}
	}

	return theme
}
