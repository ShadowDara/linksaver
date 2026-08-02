package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// GitSplitterData mirrors the gitsplitterdata dataclass.
// maxfilesize is in Megabytes.
type GitSplitterData struct {
	MaxFileSize int      `json:"maxfilesize"`
	IgnorePath  []string `json:"ignorepath"`
}

// getSplitterFolder mirrors get_splitter_folder
func getSplitterFolder(repo string) string {
	return filepath.Join(repo, ".samengine", "git-splitter")
}

// findLargeFilesInGitRepo mirrors find_large_files_in_git_repo
func findLargeFilesInGitRepo(repoPath string, config *AppConfig) ([]string, error) {
	repo, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, err
	}

	var ignorePaths []string
	if config.Git != nil {
		for _, p := range config.Git.Splitter.IgnorePath {
			ignorePaths = append(ignorePaths, filepath.ToSlash(p))
		}
	}

	maxSize := int64(config.Git.Splitter.MaxFileSize) * 1024 * 1024

	var files []string

	err = filepath.Walk(repo, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors, mirroring best-effort walk
		}
		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(repo, path)
		if err != nil {
			return nil
		}
		relSlash := filepath.ToSlash(rel)

		parts := strings.Split(relSlash, "/")
		for _, part := range parts {
			if part == ".git" || part == ".samengine" {
				return nil
			}
		}

		ignored := false
		for _, ignore := range ignorePaths {
			if relSlash == ignore || strings.HasPrefix(relSlash, ignore+"/") {
				ignored = true
				break
			}
		}
		if ignored {
			return nil
		}

		if info.Size() >= maxSize {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

// checkGit mirrors check_git: checks if executed in the git root
func checkGit(config *AppConfig) {
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		fmt.Println("git folder not found")
		fmt.Println("Please run in the git root folder")
	}
}

// checkSettings mirrors check_settings
func checkSettings(config *AppConfig) *AppConfig {
	changed := false

	if config.Git == nil {
		config.Git = &GitData{
			Submodules: []Submodules{},
			Splitter: GitSplitterData{
				MaxFileSize: 99,
				IgnorePath:  []string{},
			},
		}
		changed = true
	}

	if config.Git.Splitter.MaxFileSize == 0 && config.Git.Splitter.IgnorePath == nil {
		config.Git.Splitter = GitSplitterData{
			MaxFileSize: 99,
			IgnorePath:  []string{},
		}
		changed = true
	}

	_ = changed

	if err := saveConfig(config); err != nil {
		fmt.Println("Error saving config:", err)
	}

	return config
}

// splitterView mirrors view() in splitter.py
func splitterView(config *AppConfig) {
	repo, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Max file size: %d MB\n", config.Git.Splitter.MaxFileSize)

	files, err := findLargeFilesInGitRepo(repo, config)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("No files over maxsize found!.")
		return
	}

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		sizeMB := float64(info.Size()) / 1024 / 1024
		rel, _ := filepath.Rel(repo, file)
		fmt.Printf("%.2f MB  %s\n", sizeMB, filepath.ToSlash(rel))
	}
}

type manifestEntry struct {
	File  string   `json:"file"`
	Parts []string `json:"parts"`
}

// splitterRestore mirrors restore() in splitter.py
func splitterRestore(config *AppConfig) {
	repo, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	splitFolder := getSplitterFolder(repo)
	manifestFile := filepath.Join(splitFolder, "manifest.json")

	if _, err := os.Stat(manifestFile); os.IsNotExist(err) {
		fmt.Println("No split manifest found.")
		return
	}

	raw, err := os.ReadFile(manifestFile)
	if err != nil {
		fmt.Println("Error reading manifest:", err)
		return
	}

	var manifest []manifestEntry
	if err := json.Unmarshal(raw, &manifest); err != nil {
		fmt.Println("Error parsing manifest:", err)
		return
	}

	for _, entry := range manifest {
		original := filepath.Join(repo, entry.File)

		if err := os.MkdirAll(filepath.Dir(original), 0o755); err != nil {
			fmt.Println("Error:", err)
			continue
		}

		output, err := os.Create(original)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		for _, part := range entry.Parts {
			partFile := filepath.Join(repo, part)

			src, err := os.Open(partFile)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}

			if _, err := io.Copy(output, src); err != nil {
				fmt.Println("Error:", err)
			}
			src.Close()
		}

		output.Close()

		fmt.Printf("restored: %s\n", original)
	}

	// Cleanup
	if err := os.RemoveAll(splitFolder); err != nil {
		fmt.Println("Error cleaning up:", err)
	}
}

// splitterSplit mirrors split() in splitter.py
func splitterSplit(config *AppConfig) {
	repo, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	files, err := findLargeFilesInGitRepo(repo, config)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("No files to split found.")
		return
	}

	splitFolder := getSplitterFolder(repo)
	if err := os.MkdirAll(splitFolder, 0o755); err != nil {
		fmt.Println("Error:", err)
		return
	}

	var manifest []manifestEntry

	chunkSize := int64(config.Git.Splitter.MaxFileSize) * 1024 * 1024

	for _, file := range files {
		relative, err := filepath.Rel(repo, file)
		if err != nil {
			continue
		}

		target := filepath.Join(splitFolder, relative)
		if err := os.MkdirAll(target, 0o755); err != nil {
			fmt.Println("Error:", err)
			continue
		}

		var parts []string

		src, err := os.Open(file)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		buf := make([]byte, chunkSize)
		index := 0

		for {
			n, readErr := io.ReadFull(src, buf)
			if n > 0 {
				partName := fmt.Sprintf("%s.part%04d", filepath.Base(file), index)
				partPath := filepath.Join(target, partName)

				if err := os.WriteFile(partPath, buf[:n], 0o644); err != nil {
					fmt.Println("Error:", err)
					break
				}

				relPart, _ := filepath.Rel(repo, partPath)
				parts = append(parts, filepath.ToSlash(relPart))
				index++
			}

			if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
				break
			}
			if readErr != nil {
				fmt.Println("Error:", readErr)
				break
			}
		}

		src.Close()

		// Original entfernen
		// os.Remove(file)

		manifest = append(manifest, manifestEntry{
			File:  filepath.ToSlash(relative),
			Parts: parts,
		})

		fmt.Printf("splitted: %s\n", filepath.ToSlash(relative))
	}

	manifestJSON, err := json.MarshalIndent(manifest, "", "    ")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if err := os.WriteFile(filepath.Join(splitFolder, "manifest.json"), manifestJSON, 0o644); err != nil {
		fmt.Println("Error:", err)
	}
}

// ###############
// CLI STRUCTURE

// splitterIndex mirrors index() in splitter.py
func splitterIndex(config *AppConfig) {
	checkGit(config)
	config = checkSettings(config)

	splitterView(config)
}

// splitterCmd mirrors splitter() in splitter.py
func splitterCmd(config *AppConfig) {
	checkGit(config)
	config = checkSettings(config)

	splitterSplit(config)
}

// splitterRestoreCmd mirrors restore_splitter() in splitter.py
func splitterRestoreCmd(config *AppConfig) {
	checkGit(config)
	config = checkSettings(config)

	splitterRestore(config)
}
