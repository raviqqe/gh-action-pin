package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const version = "0.1.1"

func main() {
	showVersion := flag.Bool("version", false, "show version")
	previous := flag.Bool("previous", false, "pin to the second latest version")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	if err := run(*previous); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(previous bool) error {
	root, err := findGitRoot()
	if err != nil {
		return err
	}

	files, err := FindWorkflowFiles(root)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("no workflow files found")
	}

	resolver := newGithubResolver(previous)

	for _, file := range files {
		if err := PinWorkflowFile(file, resolver); err != nil {
			return err
		}
	}

	return nil
}

func findGitRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("finding git repository root: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}
