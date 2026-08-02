// Linksaver
// by Shadowdara
//
// This is a Go CLI program to save your links for your projects
// Read the Docs for more Infos
// https://shadowdara.github.io/docs/#/linksaver
//
// licensed under Apache license 2.0 by Shadowdara 2026
// DO NOT REMOVE THIS NOTICE !!!

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var stdinReader = bufio.NewReader(os.Stdin)

// ---------- PROMPT ----------

// prompt is a wrapper function for an input prompt
func prompt(message string) string {
	fmt.Print(message)
	line, _ := stdinReader.ReadString('\n')
	return strings.TrimSpace(line)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// ---------- INIT ----------

func initCmd() {
	fmt.Println("Init Linksaver")

	cwd, _ := os.Getwd()
	directory := filepath.Join(cwd, ".samengine")
	os.MkdirAll(directory, 0o755)

	file, _ := configPath()

	if _, err := os.Stat(file); err == nil {
		fmt.Printf("Config already exists: %s\n", file)
		return
	}

	os.WriteFile(filepath.Join(directory, "links.info.md"), []byte(""), 0o644)
	os.WriteFile(filepath.Join(directory, "links.info.txt"), []byte(""), 0o644)

	name := prompt("Projectname: ")

	config := newConfig(name)
	saveConfig(config)

	fmt.Printf("Created config at %s\n", file)
}

// ---------- ADD ----------

func add(config *AppConfig) {
	nameInput := prompt("Name (optional): ")
	authorInput := prompt("Author (optional): ")
	licenseInput := prompt("License (optional): ")
	licenseLinkInput := prompt("License Link (optional): ")

	date := time.Now().Format(time.RFC3339)

	link := Link{
		Name:         strPtr(nameInput),
		Link:         prompt("New Link: "),
		Description:  prompt("New Description: "),
		Author:       strPtr(authorInput),
		License:      strPtr(licenseInput),
		LicenseLink:  strPtr(licenseLinkInput),
		ShowInList:   prompt("Show in list? (y/n, default y): ") != "n",
		ChangeNotice: prompt("Mark as changed? (y/n, default n): ") == "y",
		Date:         &date,
	}

	config.Links = append(config.Links, link)
	saveConfig(config)

	fmt.Println("Added new link!")
}

// add4
// ersetzt Rust add2
func add4(config *AppConfig) {
	entry := prompt("Entry text: ")

	link := Link4{
		Link: entry,
		Date: time.Now().Format(time.RFC3339),
	}

	config.Links4 = append(config.Links4, link)

	saveConfig(config)

	fmt.Println("Added new entry!")
}

// add5
// ersetzt Rust add3
func add5(config *AppConfig) {
	filePath := prompt("License file: ")

	abs, err := filepath.Abs(filePath)
	if err != nil {
		fmt.Printf("Warning: '%s' does not exist.\n", filePath)
	} else if _, err := os.Stat(abs); os.IsNotExist(err) {
		fmt.Printf("Warning: '%s' does not exist.\n", filePath)
	}

	link := Link4{
		Link: filePath,
		Date: time.Now().Format(time.RFC3339),
	}

	config.Links5 = append(config.Links5, link)

	saveConfig(config)

	fmt.Println("Added license file!")
}

// ---------- OPEN LINKS ----------

func openLink(url string) {
	var err error

	switch runtime.GOOS {
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Run()
	default:
		err = exec.Command("xdg-open", url).Run()
	}

	if err != nil {
		fmt.Println("Error opening link:", err)
	}
}

func openAll(config *AppConfig) {
	fmt.Println("Opening links...")
	for _, link := range config.Links {
		openLink(link.Link)
	}
}

// ---------- MARKDOWN FORMAT ----------

func viewMarkdown(config *AppConfig) {
	file := filepath.Join(".samengine", "links.md")

	var output strings.Builder

	output.WriteString(fmt.Sprintf("# Links for %s\n\n", config.ProjectName))

	infoFile := filepath.Join(".samengine", "links.info.md")

	if data, err := os.ReadFile(infoFile); err == nil {
		output.WriteString(string(data) + "\n\n")
	} else {
		fmt.Println("Info file doesnt exist!")
	}

	output.WriteString(fmt.Sprintf("Used for %s:\n\n", config.ProjectName))

	// links
	for _, l := range config.Links {
		output.WriteString("- ")

		if l.Name != nil {
			output.WriteString(fmt.Sprintf("**%s** ", *l.Name))
		}

		output.WriteString(fmt.Sprintf("([%s](%s)) ", l.Link, l.Link))

		output.WriteString(fmt.Sprintf("- %s - ", l.Description))

		if l.Author != nil {
			output.WriteString(fmt.Sprintf("by **%s** ", *l.Author))
		}

		if l.License != nil {
			output.WriteString(fmt.Sprintf("licensed unter *%s* ", *l.License))
		}

		if l.LicenseLink != nil {
			output.WriteString(fmt.Sprintf("([%s](%s)) ", *l.LicenseLink, *l.LicenseLink))
		}

		if l.ChangeNotice {
			output.WriteString("- *(changes were made)*")
		}

		if l.Date != nil {
			output.WriteString(fmt.Sprintf(" - *(saved at date: %s)*", *l.Date))
		}

		output.WriteString("\n")
	}

	// links2
	for _, l := range config.Links2 {
		output.WriteString(fmt.Sprintf("- %s\n", l))
	}

	// links4
	for _, l := range config.Links4 {
		output.WriteString(fmt.Sprintf("- %s saved at date: **%s**\n", l.Link, l.Date))
	}

	// package lock
	for _, item := range config.LinksPkgLock {
		output.WriteString(fmt.Sprintf(
			"- used **%s** version %s licensed under **%v** - *[Link](%s)*\n",
			item.Name, item.Version, item.License, item.Link,
		))
	}

	// cargo lock
	for _, item := range config.LinksCargoLock {
		output.WriteString(fmt.Sprintf(
			"- used **%s** version %s licensed under **%v** - *[Link](%s)*\n",
			item.Name, item.Version, item.License, item.Link,
		))
	}

	// links3
	for _, licenseFile := range config.Links3 {
		if content, err := os.ReadFile(licenseFile); err == nil {
			output.WriteString(fmt.Sprintf(`
---

**license content from file: %s**

`+"```"+`
%s
`+"```"+`

`, licenseFile, string(content)))
		} else {
			fmt.Printf("Warning: License file '%s' does not exist.\n", licenseFile)
		}
	}

	// links5
	for _, item := range config.Links5 {
		if content, err := os.ReadFile(item.Link); err == nil {
			output.WriteString(fmt.Sprintf(`
---

**license content from file: %s** at date: *%s*

`+"```"+`
%s
`+"```"+`

`, item.Link, item.Date, string(content)))
		} else {
			fmt.Printf("Warning: License file '%s' does not exist.\n", item.Link)
		}
	}

	output.WriteString(`
---

*File generated by linksaver from s2* - [More Infos](https://shadowdara.wordpress.com/2026/06/30/minisite-a-site-in-only-one-html-file/)
`)

	os.WriteFile(file, []byte(output.String()), 0o644)

	fmt.Println(`
Created File - Use parseMarkdown from samengine to make it into a nice html file.

npm i samengine
npx samengine markdown .samengine/links.md
`)
}

// ---------- TXT FORMAT ----------

// viewx converts the licenses in the JSON file into a TXT File
func viewx(config *AppConfig) {
	file := filepath.Join(".samengine", "links.txt")

	var output strings.Builder

	output.WriteString(fmt.Sprintf("Links for %s\n\n", config.ProjectName))

	infoFile := filepath.Join(".samengine", "links.info.txt")

	if data, err := os.ReadFile(infoFile); err == nil {
		output.WriteString(string(data) + "\n\n")
	} else {
		fmt.Println("Info file doesnt exist!")
	}

	output.WriteString(fmt.Sprintf("Used for %s:\n\n", config.ProjectName))

	// links
	for _, l := range config.Links {
		output.WriteString("- ")

		if l.Name != nil {
			output.WriteString(*l.Name)
		}

		output.WriteString(fmt.Sprintf(" (%s) ", l.Link))

		output.WriteString(fmt.Sprintf("- %s - ", l.Description))

		if l.Author != nil {
			output.WriteString(fmt.Sprintf("by %s ", *l.Author))
		}

		if l.License != nil {
			output.WriteString(fmt.Sprintf("licensed unter %s ", *l.License))
		}

		if l.LicenseLink != nil {
			output.WriteString(fmt.Sprintf("(%s) ", *l.LicenseLink))
		}

		if l.ChangeNotice {
			output.WriteString("- (changes were made)")
		}

		if l.Date != nil {
			output.WriteString(fmt.Sprintf(" - (saved at date: %s)", *l.Date))
		}

		output.WriteString("\n")
	}

	// links2
	for _, l := range config.Links2 {
		output.WriteString(fmt.Sprintf("- %s\n", l))
	}

	// links4
	for _, l := range config.Links4 {
		output.WriteString(fmt.Sprintf("- %s saved at date: %s\n", l.Link, l.Date))
	}

	// links3
	for _, licenseFile := range config.Links3 {
		if content, err := os.ReadFile(licenseFile); err == nil {
			output.WriteString(fmt.Sprintf(`
license content from file: %s

%s

`, licenseFile, string(content)))
		} else {
			fmt.Printf("Warning: License file '%s' does not exist.\n", licenseFile)
		}
	}

	// links5
	for _, item := range config.Links5 {
		if content, err := os.ReadFile(item.Link); err == nil {
			output.WriteString(fmt.Sprintf(`
license content from file: %s at date: %s

%s

`, item.Link, item.Date, string(content)))
		} else {
			fmt.Printf("Warning: License file '%s' does not exist.\n", item.Link)
		}
	}

	output.WriteString(`
---

File generated by linksaver from s2 - https://shadowdara.wordpress.com/2026/06/30/minisite-a-site-in-only-one-html-file/
`)

	os.WriteFile(file, []byte(output.String()), 0o644)
}

// ---------- LIST ----------

func listLinks(config *AppConfig) {
	fmt.Println("\nCredits:")
	fmt.Println()

	for _, l := range config.Links {
		if !l.ShowInList {
			continue
		}

		name := ""
		if l.Name != nil {
			name = *l.Name
		}
		author := ""
		if l.Author != nil {
			author = *l.Author
		}
		license := ""
		if l.License != nil {
			license = *l.License
		}
		licenselink := ""
		if l.LicenseLink != nil {
			licenselink = *l.LicenseLink
		}
		changed := ""
		if l.ChangeNotice {
			changed = " (changes were made)"
		}

		fmt.Printf(
			"\"%s\" (%s) by %s is licensed under %s (%s)%s\n",
			name, l.Link, author, license, licenselink, changed,
		)
	}

	for _, entry := range config.Links2 {
		fmt.Println(entry)
	}
}

// ---------- ADD PACKAGE LOCK LICENSES ----------

type npmPackageJSON struct {
	License  interface{}   `json:"license,omitempty"`
	Licenses []interface{} `json:"licenses,omitempty"`
}

func readNpmLicense(pkgPath string) string {
	data, err := os.ReadFile(filepath.Join(pkgPath, "package.json"))
	if err != nil {
		return "UNKNOWN"
	}

	var pkgJSON npmPackageJSON
	if err := json.Unmarshal(data, &pkgJSON); err != nil {
		return "UNKNOWN"
	}

	switch v := pkgJSON.License.(type) {
	case string:
		return v
	case map[string]interface{}:
		if t, ok := v["type"].(string); ok {
			return t
		}
	}

	if pkgJSON.Licenses != nil {
		var parts []string
		for _, x := range pkgJSON.Licenses {
			switch v := x.(type) {
			case string:
				parts = append(parts, v)
			case map[string]interface{}:
				if t, ok := v["type"].(string); ok {
					parts = append(parts, t)
				} else {
					parts = append(parts, "")
				}
			}
		}
		return strings.Join(parts, ", ")
	}

	return "UNKNOWN"
}

func addPkgLock(config *AppConfig) {
	cwd, _ := os.Getwd()
	lockFile := filepath.Join(cwd, "package-lock.json")

	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		fmt.Println("package-lock.json not found")
		return
	}

	nodeModules := filepath.Join(cwd, "node_modules")
	if _, err := os.Stat(nodeModules); os.IsNotExist(err) {
		fmt.Println("node_modules not found. Run npm install first.")
		return
	}

	raw, err := os.ReadFile(lockFile)
	if err != nil {
		fmt.Println("Error reading package-lock.json:", err)
		return
	}

	var lock map[string]json.RawMessage
	if err := json.Unmarshal(raw, &lock); err != nil {
		fmt.Println("Error parsing package-lock.json:", err)
		return
	}

	var packages []PackageInfo

	nodeModulesRe := regexp.MustCompile(`^node_modules/`)

	// package-lock v2 / v3
	if pkgsRaw, ok := lock["packages"]; ok {
		var pkgs map[string]struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}
		if err := json.Unmarshal(pkgsRaw, &pkgs); err == nil {
			for key, value := range pkgs {
				if key == "" {
					continue
				}

				packagePath := filepath.Join(cwd, key)

				name := value.Name
				if name == "" {
					name = nodeModulesRe.ReplaceAllString(key, "")
				}

				packages = append(packages, PackageInfo{
					Name:    name,
					Version: value.Version,
					License: readNpmLicense(packagePath),
					Link:    fmt.Sprintf("https://www.npmjs.com/package/%s", name),
					Date:    time.Now().Format(time.RFC3339),
				})
			}
		}
	}

	config.LinksPkgLock = packages

	saveConfig(config)

	fmt.Printf("Added %d packages from package-lock.json\n", len(packages))
}

// ---------- ADD CARGO LOCK LICENSES ----------

func findCargoToml(cargoHome, name, version string) string {
	entries, err := os.ReadDir(cargoHome)
	if err != nil {
		return ""
	}

	for _, registry := range entries {
		cargoToml := filepath.Join(cargoHome, registry.Name(), fmt.Sprintf("%s-%s", name, version), "Cargo.toml")
		if _, err := os.Stat(cargoToml); err == nil {
			return cargoToml
		}
	}

	return ""
}

var cargoLicenseRe = regexp.MustCompile(`(?m)^\s*license\s*=\s*"([^"]+)"`)
var cargoLicenseFileRe = regexp.MustCompile(`(?m)^\s*license-file\s*=\s*"([^"]+)"`)
var cargoNameRe = regexp.MustCompile(`(?m)^\s*name\s*=\s*"([^"]+)"`)
var cargoVersionRe = regexp.MustCompile(`(?m)^\s*version\s*=\s*"([^"]+)"`)

func readCargoLicense(cargoToml string) string {
	content, err := os.ReadFile(cargoToml)
	if err != nil {
		return "UNKNOWN"
	}

	if m := cargoLicenseRe.FindStringSubmatch(string(content)); m != nil {
		return m[1]
	}

	if cargoLicenseFileRe.MatchString(string(content)) {
		return "SEE LICENSE FILE"
	}

	return "UNKNOWN"
}

// addCargoLock adds all the licenses from Cargo packages
func addCargoLock(config *AppConfig) {
	cwd, _ := os.Getwd()
	lockFile := filepath.Join(cwd, "Cargo.lock")

	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		fmt.Println("Cargo.lock not found")
		return
	}

	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		fmt.Println("Home directory not found")
		return
	}

	cargoHome := filepath.Join(home, ".cargo", "registry", "src")
	if _, err := os.Stat(cargoHome); os.IsNotExist(err) {
		fmt.Println("Cargo registry not found.")
		fmt.Println("Run: cargo fetch")
		return
	}

	lockBytes, err := os.ReadFile(lockFile)
	if err != nil {
		fmt.Println("Error reading Cargo.lock:", err)
		return
	}
	lock := string(lockBytes)

	var packages []PackageInfo

	blocks := strings.Split(lock, "[[package]]")

	seen := make(map[string]bool)

	for _, block := range blocks {
		nameMatch := cargoNameRe.FindStringSubmatch(block)
		versionMatch := cargoVersionRe.FindStringSubmatch(block)

		if nameMatch == nil || versionMatch == nil {
			continue
		}

		name := nameMatch[1]
		version := versionMatch[1]

		identifier := fmt.Sprintf("%s@%s", name, version)

		if seen[identifier] {
			continue
		}
		seen[identifier] = true

		cargoToml := findCargoToml(cargoHome, name, version)

		license := "UNKNOWN"
		if cargoToml != "" {
			license = readCargoLicense(cargoToml)
		}

		packages = append(packages, PackageInfo{
			Name:    name,
			Version: version,
			License: license,
			Link:    fmt.Sprintf("https://crates.io/crates/%s", name),
			Date:    time.Now().Format(time.RFC3339),
		})
	}

	config.LinksCargoLock = packages

	saveConfig(config)

	fmt.Printf("Added %d crates from Cargo.lock\n", len(packages))
}

// addGitSubmodule adds a git submodule to the data file
func addGitSubmodule(config *AppConfig) {
	desc := prompt("Description: ")
	dirrr := prompt("Dir (where git clone is executed): ")
	clonedir := prompt("The name for the repo dir: ")
	repolink := prompt("repo link: ")
	repocommit := prompt("repo commit: ")
	branch := prompt("Repo Branch (empty for the main branch): ")

	var branchPtr *string
	if branch != "" {
		branchPtr = &branch
	}

	module := Submodules{
		Dir:        dirrr,
		Repolink:   repolink,
		Repocommit: repocommit,
		Clonedir:   clonedir,
		Desc:       desc,
		Branch:     branchPtr,
	}

	if config.Git == nil {
		config.Git = &GitData{
			Submodules: []Submodules{},
			Splitter: GitSplitterData{
				MaxFileSize: 99,
				IgnorePath:  []string{},
			},
		}
	}

	// add to the config
	config.Git.Submodules = append(config.Git.Submodules, module)

	// DONT FORGET SAVING!
	saveConfig(config)

	fmt.Println("Added new submodule")
}

// cloneGitSubmodules clones the git submodules
func cloneGitSubmodules(config *AppConfig) {
	fmt.Println("Cloning depencies")

	oldPath, _ := os.Getwd()

	if config.Git == nil {
		fmt.Println("git option is None!")
		return
	}

	for _, e := range config.Git.Submodules {
		// Reset the path
		os.Chdir(oldPath)

		// Print Infos
		fmt.Println(e.Desc)

		// Create dir
		cwd, _ := os.Getwd()
		os.MkdirAll(cwd+"/"+e.Dir, 0o755)

		// Change the execution directory
		os.Chdir(cwd + "/" + e.Dir)

		// Clone
		var cloneCommand string

		// With a different branch here
		if e.Branch != nil {
			cloneCommand = fmt.Sprintf(
				`git clone --recursive --branch "%s" "%s" "%s"`,
				*e.Branch, e.Repolink, e.Clonedir,
			)
		} else {
			cloneCommand = fmt.Sprintf(
				`git clone --recursive "%s" "%s"`,
				e.Repolink, e.Clonedir,
			)
		}

		runShell(cloneCommand)

		// Change dir to the clone dir
		cwd, _ = os.Getwd()
		os.Chdir(cwd + "/" + e.Clonedir)

		// Run git checkout for that commit
		checkoutCommand := "git checkout " + e.Repocommit
		runShell(checkoutCommand)

		runShell("git submodule update --init --recursive")

		// Clone l2 dependencies
		runShell("l2 clonesubm")

		fmt.Printf("Cloned %s successfuly!\n", e.Clonedir)
	}

	fmt.Println("Finished cloning every submodule!")
}

func runShell(command string) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	_ = cmd.Run()
}

// ---------- HELP ----------

func banner() {
	fmt.Println(`
██╗     ██╗███╗   ██╗██╗  ██╗███████╗ █████╗ ██╗   ██╗███████╗██████╗
██║     ██║████╗  ██║██║ ██╔╝██╔════╝██╔══██╗██║   ██║██╔════╝██╔══██╗
██║     ██║██╔██╗ ██║█████╔╝ ███████╗███████║██║   ██║█████╗  ██████╔╝
██║     ██║██║╚██╗██║██╔═██╗ ╚════██║██╔══██║╚██╗ ██╔╝██╔══╝  ██╔══██╗
███████╗██║██║ ╚████║██║  ██╗███████║██║  ██║ ╚████╔╝ ███████╗██║  ██║
╚══════╝╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝  ╚═══╝  ╚══════╝╚═╝  ╚═╝
`)
}

func helpCmd() {
	banner()
	fmt.Printf(`by %sshadowdara%s Version %s%s%s

=== Commands ===

help            show this message
init            create config
add             add link
add2            add entry (text only)
add3            add license file
view            formats links into Markdown
viewx           formats links into TXT
list            list links
addpkg          add links from a package lock file
addcargo        add links from a cargo lock file
open            open all links
info            get more infos about the programm
addsubmodule    add a git submodule to the data (more infos in the docs)
clonesubm, %sc%s    clone the git submodules (requires git)
gitsplit        Split files in repo which are to big for git
gitrestore      restore the splitted files
gitview         View the files which are to big for git
s               a little status info with gitview and git status
`, YELLOW, END, GREEN, Version, END, PURPLE, END)
}

// info generates an Info on how to use the Program (TODO)
func infoCmd() {
}

// ---------- MENU ----------

type menuCommand struct {
	name   string
	arg    string // empty string is a valid arg ("open"); use hasArg to detect Exit
	hasArg bool
}

// menu shows a selection menu in the command line so the user doesn't
// have to select the options via CLI args
func menu() string {
	commands := []menuCommand{
		{"Open all links", "", true},
		{"Init", "init", true},
		{"Add link", "add", true},
		{"Add text entry", "add2", true},
		{"Add license file", "add3", true},
		{"Generate Markdown", "view", true},
		{"Generate TXT", "viewx", true},
		{"List credits", "list", true},
		{"Import package-lock.json", "addpkg", true},
		{"Import Cargo.lock", "addcargo", true},
		{"Help", "help", true},
		{"add Git Submodule", "addsubmodule", true},
		{"Clone git submodules", "clonesubm", true},
		{"Split to big files for git", "gitsplit", true},
		{"Restore the files which are to big", "gitrestore", true},
		{"View files which are to big", "gitview", true},
		{"Exit", "", false},
	}

	fmt.Println("\n=== Linksaver ===\n")

	for i, c := range commands {
		fmt.Printf("%d. %s\n", i+1, c.name)
	}

	for {
		choiceStr := prompt("\nSelect: ")
		choice, err := strconv.Atoi(choiceStr)

		if err == nil && choice >= 1 && choice <= len(commands) {
			cmd := commands[choice-1]
			if !cmd.hasArg {
				return ""
			}
			return cmd.arg
		}

		fmt.Println("Invalid selection.")
	}
}

// ---------- STATUS ----------

// statusCmd displays an easy stats menu instead of git status
func statusCmd() {
	runShell("git status")
	runShell("l2 gitview")
}

// ---------- EXECUTE ----------

func execute(arg string, config *AppConfig) {
	switch arg {
	case "help", "-h", "--help", "h":
		helpCmd()
		return
	case "info":
		infoCmd()
		return
	case "init":
		initCmd()
		return
	case "add":
		add(config)
	case "add2":
		add4(config)
	case "add3":
		add5(config)
	case "view":
		viewMarkdown(config)
	case "viewx":
		viewx(config)
	case "list":
		listLinks(config)
	case "addpkg":
		addPkgLock(config)
	case "addcargo":
		addCargoLock(config)
	case "addsubmodule":
		addGitSubmodule(config)
	case "clonesubm", "c":
		cloneGitSubmodules(config)
	case "gitsplit":
		splitterCmd(config)
	case "gitrestore":
		splitterRestoreCmd(config)
	case "gitview":
		splitterIndex(config)
	case "open":
		openAll(config)
	case "s":
		statusCmd()
	default:
		fmt.Println("Linksaver: Argument not found!")
	}
}

// ---------- MAIN ----------

func main() {
	// Try loading the Config
	config, err := loadConfig()

	if err != nil {
		// Option to create a new config when no one exists
		if len(os.Args) > 1 {
			if os.Args[1] == "init" {
				execute("init", newConfig("temp"))
				return
			}

			switch os.Args[1] {
			case "help", "-h", "--help", "h":
				helpCmd()
				return
			}
		}

		// When a Config Error Appears
		banner()
		fmt.Println("Linksaver: Config Error:", err)
		fmt.Println("Run 'init' first or run with help!")

		time.Sleep(2 * time.Second)
		os.Exit(1)
	}

	if config.Settings != nil && config.Settings.SelectMenu {
		// Select via a cli selector
		arg := menu()

		// Execute the selection
		execute(arg, config)

		// finish
		return
	}

	// More than one argument
	// then run linksaver in cli mode
	if len(os.Args) > 1 {
		// get first arg after the program name
		arg := os.Args[1]

		execute(arg, config)
	} else {
		fmt.Println("Linksaver: run with one argument of help!")
	}
}
