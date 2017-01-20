package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/ungerik/go-fs"
)

var config struct {
	GoPath      string `json:"GOPATH,omitempty"`
	GitHub      string `json:"GitHub,omitempty"`
	ShowDetails bool   `json:"ShowDetails,omitempty"`
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	home, err = homedir.Expand(home)
	if err != nil {
		panic(err)
	}

	configPath := filepath.Join(cwd, ".gitstat")
	if !fs.Exists(configPath) {
		configPath = filepath.Join(home, ".gitstat")
	}

	if fs.Exists(configPath) {
		err = fs.ReadJSON(fs.GetFile(configPath), &config)
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println(".gitstat not found")
	}

	if config.GoPath == "" {
		config.GoPath = os.Getenv("GOPATH")
		if config.GoPath == "" {
			config.GoPath = filepath.Join(home, "go")
		}
	}

	flag.StringVar(&config.GoPath, "GOPATH", config.GoPath, "change if you don't want to use the system GOPATH")
	flag.StringVar(&config.GitHub, "GitHub", config.GitHub, "GitHub username")
	flag.BoolVar(&config.ShowDetails, "ShowDetails", config.ShowDetails, "shows full git status output")
	flag.Parse()

	githubProjectPath := filepath.Join(config.GoPath, "src", "github.com", config.GitHub)
	if !fs.IsDir(githubProjectPath) {
		panic("GitHub project path not found: " + githubProjectPath)
	}

	var projectDirs []fs.File
	err = fs.ListDir(githubProjectPath, func(file fs.File) error {
		if file.IsDir() {
			projectDirs = append(projectDirs, file)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("====================================================")
	fmt.Println()

	for _, projectDir := range projectDirs {
		err = os.Chdir(projectDir.Path())
		if err != nil {
			panic(err)
		}

		out, err := exec.Command("git", "status").CombinedOutput()
		if err != nil {
			panic(err)
		}

		output := string(out)
		if strings.Contains(output, "nothing to commit, working tree clean") {
			continue
		}

		hasUntrackedFiles := strings.Contains(output, "Untracked files:")
		hasNotStagedChanges := strings.Contains(output, "Changes not staged for commit:")

		if config.ShowDetails {
			fmt.Println(projectDir.Path())
			fmt.Println(string(out))
			fmt.Println()
		} else {
			fmt.Print(projectDir.Path())
			if hasUntrackedFiles {
				fmt.Print(" -> Untracked files")
			}
			if hasNotStagedChanges {
				fmt.Print(" -> Changes not staged for commit")
			}
			if !hasUntrackedFiles && !hasNotStagedChanges {
				fmt.Print(" -> UNKNOWN STATE")
			}
			fmt.Println()
		}
	}

	os.Chdir(cwd)
}
