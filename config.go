package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ---------- DATACLASSES ----------

// Submodules mirrors the Submodules dataclass
type Submodules struct {
	Desc       string  `json:"desc"`
	Clonedir   string  `json:"clonedir"`
	Dir        string  `json:"dir"`
	Repolink   string  `json:"repolink"`
	Repocommit string  `json:"repocommit"`
	Branch     *string `json:"branch,omitempty"`
}

// GitData mirrors the GitData dataclass
type GitData struct {
	Submodules []Submodules    `json:"submodules"`
	Splitter   GitSplitterData `json:"splitter"`
}

// Settings mirrors the Settings dataclass
type Settings struct {
	SelectMenu bool `json:"selectmenu"`
}

// Link mirrors the Link dataclass
type Link struct {
	Link         string  `json:"link"`
	Description  string  `json:"description"`
	Name         *string `json:"name,omitempty"`
	License      *string `json:"license,omitempty"`
	Author       *string `json:"author,omitempty"`
	LicenseLink  *string `json:"licenselink,omitempty"`
	ShowInList   bool    `json:"showinlist"`
	ChangeNotice bool    `json:"changenotice"`
	Date         *string `json:"date,omitempty"`
}

// PackageInfo mirrors the PackageInfo dataclass (used for npm/cargo packages)
type PackageInfo struct {
	Name    string      `json:"name"`
	Link    string      `json:"link"`
	Version string      `json:"version"`
	Date    string      `json:"date"`
	License interface{} `json:"license,omitempty"` // string or []string
}

// Link4 mirrors the Link4 dataclass (used for links4/links5)
type Link4 struct {
	Link string `json:"link"`
	Date string `json:"date"`
}

// AppConfig mirrors the AppConfig dataclass
type AppConfig struct {
	ProjectName string `json:"projectname"`
	Pretty      bool   `json:"pretty"`

	Schema *string `json:"$schema,omitempty"`

	Links  []Link   `json:"links"`
	Links2 []string `json:"links2"`
	Links3 []string `json:"links3"`
	Links4 []Link4  `json:"links4"`
	Links5 []Link4  `json:"links5"`

	LinksPkgLock   []PackageInfo `json:"linkspkglock"`
	LinksCargoLock []PackageInfo `json:"linkscargolock"`

	Settings *Settings `json:"settings,omitempty"`

	Git *GitData `json:"git,omitempty"`

	Note *string `json:"note,omitempty"`
}

// ---------- CONSTANTS ----------

const NOTE = "This file was generated with linksaver by Shadowdara for the samengine project. see https://shadowara.github.io/docs#/linksaver or or https://github.com/shadowdara/l2 for more infos"

const SCHEMA_URL = "https://raw.githubusercontent.com/ShadowDara/l2/refs/heads/master/shema.json"

// ---------- PATH ----------

// configPath returns the path to the config file
func configPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, "linksaver.json"), nil
	// return filepath.Join(cwd, ".samengine", "linksaver.json"), nil
}

// saveConfig saves the AppConfig to disk as JSON
func saveConfig(config *AppConfig) error {
	file, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}

	var text []byte

	if config.Pretty {
		text, err = json.MarshalIndent(config, "", "    ")
	} else {
		text, err = json.Marshal(config)
	}
	if err != nil {
		return err
	}

	return os.WriteFile(file, text, 0o644)
}

// ---------- CONFIG ----------

// newSettingsConfig returns a fresh Settings struct
func newSettingsConfig() *Settings {
	return &Settings{
		SelectMenu: false,
	}
}

// newConfig creates a new AppConfig with default values
func newConfig(name string) *AppConfig {
	schema := SCHEMA_URL
	note := NOTE

	return &AppConfig{
		ProjectName:    name,
		Schema:         &schema,
		Pretty:         true,
		Links:          []Link{},
		Links2:         []string{},
		Links3:         []string{},
		Links4:         []Link4{},
		Links5:         []Link4{},
		LinksPkgLock:   []PackageInfo{},
		LinksCargoLock: []PackageInfo{},
		Settings:       newSettingsConfig(),
		Note:           &note,
	}
}

// loadConfig loads the linksaver config from disk
func loadConfig() (*AppConfig, error) {
	file, err := configPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil, errors.New("config not found")
	}

	raw, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	// Use a generic map first, mirroring the Python dict-based loading
	var data map[string]json.RawMessage
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}

	var projectName string
	if pnRaw, ok := data["projectname"]; ok {
		_ = json.Unmarshal(pnRaw, &projectName)
	}
	if projectName == "" {
		return nil, errors.New("projectname must be set")
	}

	schema := SCHEMA_URL
	if sRaw, ok := data["$schema"]; ok {
		var s string
		if err := json.Unmarshal(sRaw, &s); err == nil && s != "" {
			schema = s
		}
	}

	note := NOTE

	config := &AppConfig{
		ProjectName: projectName,
		Pretty:      true,
		Schema:      &schema,
		Note:        &note,
	}

	if pRaw, ok := data["pretty"]; ok {
		_ = json.Unmarshal(pRaw, &config.Pretty)
	}

	config.Links = []Link{}
	if lRaw, ok := data["links"]; ok {
		_ = json.Unmarshal(lRaw, &config.Links)
	}

	config.Links2 = []string{}
	if lRaw, ok := data["links2"]; ok {
		_ = json.Unmarshal(lRaw, &config.Links2)
	}

	config.Links3 = []string{}
	if lRaw, ok := data["links3"]; ok {
		_ = json.Unmarshal(lRaw, &config.Links3)
	}

	config.Links4 = []Link4{}
	if lRaw, ok := data["links4"]; ok {
		_ = json.Unmarshal(lRaw, &config.Links4)
	}

	config.Links5 = []Link4{}
	if lRaw, ok := data["links5"]; ok {
		_ = json.Unmarshal(lRaw, &config.Links5)
	}

	config.LinksPkgLock = []PackageInfo{}
	if lRaw, ok := data["linkspkglock"]; ok {
		_ = json.Unmarshal(lRaw, &config.LinksPkgLock)
	}

	config.LinksCargoLock = []PackageInfo{}
	if lRaw, ok := data["linkscargolock"]; ok {
		_ = json.Unmarshal(lRaw, &config.LinksCargoLock)
	}

	if sRaw, ok := data["settings"]; ok {
		var settings Settings
		if err := json.Unmarshal(sRaw, &settings); err == nil {
			config.Settings = &settings
		}
	}
	if config.Settings == nil {
		config.Settings = newSettingsConfig()
	}

	// git
	var gitRaw map[string]json.RawMessage
	if gRaw, ok := data["git"]; ok {
		_ = json.Unmarshal(gRaw, &gitRaw)
	}

	var submodules []Submodules
	if sRaw, ok := gitRaw["submodules"]; ok {
		_ = json.Unmarshal(sRaw, &submodules)
	}
	if submodules == nil {
		submodules = []Submodules{}
	}

	splitterData := GitSplitterData{
		MaxFileSize: 99,
		IgnorePath:  []string{},
	}
	if spRaw, ok := gitRaw["splitter"]; ok {
		var sp map[string]json.RawMessage
		if err := json.Unmarshal(spRaw, &sp); err == nil {
			if mfsRaw, ok := sp["maxfilesize"]; ok {
				_ = json.Unmarshal(mfsRaw, &splitterData.MaxFileSize)
			}
			if ipRaw, ok := sp["ignorepath"]; ok {
				_ = json.Unmarshal(ipRaw, &splitterData.IgnorePath)
			}
		}
	}

	config.Git = &GitData{
		Submodules: submodules,
		Splitter:   splitterData,
	}

	return config, nil
}

// helper used elsewhere for error formatting, mirrors python's generic Exception messages
func configErrorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}
