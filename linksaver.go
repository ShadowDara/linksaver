// linksaver.go
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Link struct {
	Name        string `json:"name,omitempty"`
	Link        string `json:"link"`
	Description string `json:"description"`
	License     string `json:"license,omitempty"`
	Author      string `json:"author,omitempty"`
	LicenseLink string `json:"licenselink,omitempty"`
}

type AppConfig struct {
	ProjectName string `json:"projectname"`
	Pretty      bool   `json:"pretty"`
	Links       []Link `json:"links"`
}

var baseDir, _ = os.Getwd()
var configPath = filepath.Join(baseDir, ".samengine", "linksaver.json")

// ---------- Utils ----------

func newAppConfig(name string) AppConfig {
	return AppConfig{
		ProjectName: name,
		Pretty:      true,
		Links:       []Link{},
	}
}

func saveConfig(config AppConfig) error {
	var data []byte
	var err error

	if config.Pretty {
		data, err = json.MarshalIndent(config, "", "    ")
	} else {
		data, err = json.Marshal(config)
	}

	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func loadConfig() (AppConfig, error) {
	var config AppConfig

	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		return config, fmt.Errorf("config not found")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("invalid json: %w", err)
	}

	if config.ProjectName == "" {
		return config, fmt.Errorf("projectname must be set")
	}

	return config, nil
}

func prompt(question string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(question)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

// ---------- Commands ----------

func initConfig() {
	fmt.Println("Init Linksaver")

	os.MkdirAll(".samengine", os.ModePerm)

	if _, err := os.Stat(configPath); err == nil {
		fmt.Println("Config already exists:", configPath)
		return
	}

	projectName := prompt("Projectname: ")

	config := newAppConfig(projectName)

	err := saveConfig(config)
	if err != nil {
		fmt.Println("Error saving config:", err)
		return
	}

	fmt.Println("Created config at", configPath)
}

func addLink(config *AppConfig) {
	name := prompt("Name (optional): ")
	link := prompt("New Link: ")
	desc := prompt("New Description: ")
	author := prompt("Author (optional): ")
	license := prompt("License (optional): ")
	licenseLink := prompt("License Link (optional): ")

	newLink := Link{
		Name:        name,
		Link:        link,
		Description: desc,
		Author:      author,
		License:     license,
		LicenseLink: licenseLink,
	}

	config.Links = append(config.Links, newLink)

	err := saveConfig(*config)
	if err != nil {
		fmt.Println("Error saving:", err)
		return
	}

	fmt.Println("Added new link!")
}

func viewLinks(config AppConfig) {
	for _, l := range config.Links {
		fmt.Printf("[%s] ", l.Link)

		if l.Name != "" {
			fmt.Printf("Name: %s | ", l.Name)
		}

		fmt.Printf("Desc: %s | ", l.Description)

		if l.Author != "" {
			fmt.Printf("Author: %s | ", l.Author)
		}

		if l.License != "" {
			fmt.Printf("License: %s | ", l.License)
		}

		if l.LicenseLink != "" {
			fmt.Printf("License URL: %s", l.LicenseLink)
		}

		fmt.Println()
	}
}

func listLinks(config AppConfig) {
	for _, l := range config.Links {
		fmt.Printf("\"%s\" (%s) by %s is licensed under %s (%s)\n",
			l.Name, l.Link, l.Author, l.License, l.LicenseLink)
	}
}

func openLink(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}

	err := cmd.Start()
	if err != nil {
		fmt.Println("Error opening:", err)
	}
}

func openLinks(config AppConfig) {
	fmt.Println("Opening links...")
	for _, l := range config.Links {
		openLink(l.Link)
	}
}

func printHelp() {
	fmt.Println(`
LINKSAVER CLI

Commands:
    help    show this message
    init    create config
    add     add link
    view    view links
    list    list links
    (none)  open all links
`)
}

// ---------- Main ----------

func main() {
	args := os.Args[1:]

	if len(args) > 0 {
		switch args[0] {
		case "help", "-h", "--help":
			printHelp()
			return
		case "init":
			initConfig()
			return
		}
	}

	config, err := loadConfig()
	if err != nil {
		fmt.Println("Linksaver")
		fmt.Println("Config Error:", err)
		fmt.Println("Run 'init' first.")
		os.Exit(1)
	}

	if len(args) > 0 {
		switch args[0] {
		case "add":
			addLink(&config)
			return
		case "view":
			viewLinks(config)
			return
		case "list":
			listLinks(config)
			return
		}
	}

	openLinks(config)
}
