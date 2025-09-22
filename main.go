// gvs is a simple Go version manager.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const gvsDir = ".gvs"

var ErrVersionNotFound = errors.New("version not found")

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "completion":
		handleCompletion()
	case "list-remote":
		listRemoteCmd := flag.NewFlagSet("list-remote", flag.ExitOnError)

		limit := listRemoteCmd.Int("limit", 10, "Limit the number of versions returned")
		if err := listRemoteCmd.Parse(os.Args[2:]); err != nil {
			log.Fatalf("failed to parse flags: %v", err)
		}

		err := listRemoteVersions(*limit)
		if err != nil {
			log.Fatal(err)
		}
	case "list":
		err := listLocalVersions()
		if err != nil {
			log.Fatal(err)
		}
	case "install":
		if len(os.Args) < 3 {
			fmt.Println("Please specify a version to install.")
			printUsage()
			os.Exit(1)
		}

		err := installVersion(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
	case "use":
		if len(os.Args) < 3 {
			fmt.Println("Please specify a version to use.")
			printUsage()
			os.Exit(1)
		}

		err := useVersion(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
	case "remove":
		if len(os.Args) < 3 {
			fmt.Println("Please specify a version to remove.")
			printUsage()
			os.Exit(1)
		}

		err := removeVersion(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
	default:
		printUsage()
		os.Exit(1)
	}
}

func listRemoteVersions(limit int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://go.dev/dl/", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get Go versions: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	re := regexp.MustCompile(`go(\d+\.\d+(\.\d+)?)\.src\.tar\.gz`)
	matches := re.FindAllStringSubmatch(string(body), -1)

	versions := make(map[string]bool)
	count := 0

	for _, match := range matches {
		if !versions[match[1]] {
			fmt.Println(match[1])
			versions[match[1]] = true
			count++

			if limit > 0 && count >= limit {
				break
			}
		}
	}

	return nil
}

func listLocalVersions() error {
	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	versionsDir := filepath.Join(usr.HomeDir, "sdk")
	if _, err := os.Stat(versionsDir); os.IsNotExist(err) {
		fmt.Println("No Go versions installed.")
		return nil
	}

	dirs, err := os.ReadDir(versionsDir)
	if err != nil {
		return fmt.Errorf("failed to read versions directory: %w", err)
	}

	for _, dir := range dirs {
		if dir.IsDir() {
			fmt.Println(dir.Name())
		}
	}

	return nil
}

func installVersion(version string) error {
	bareVersion := strings.TrimPrefix(version, "go")
	fmt.Printf("Installing Go %s...\n", bareVersion)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// #nosec G204
	installCmd := exec.CommandContext(ctx, "go", "install", fmt.Sprintf("golang.org/dl/go%s@latest", bareVersion))
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr

	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install Go version downloader: %w", err)
	}

	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	goDownloader := filepath.Join(usr.HomeDir, "go", "bin", "go"+bareVersion)

	// #nosec G204
	downloadCmd := exec.CommandContext(ctx, goDownloader, "download")
	downloadCmd.Stdout = os.Stdout
	downloadCmd.Stderr = os.Stderr

	if err := downloadCmd.Run(); err != nil {
		return fmt.Errorf("failed to download Go version: %w", err)
	}

	fmt.Printf("Go %s installed successfully.\n", bareVersion)

	return nil
}

func useVersion(version string) error {
	installed, err := isVersionInstalled(version)
	if err != nil {
		return err
	}

	if !installed {
		return fmt.Errorf("%w: %s", ErrVersionNotFound, version)
	}

	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	binDir := filepath.Join(usr.HomeDir, gvsDir, "bin")
	if err := os.MkdirAll(binDir, 0o750); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	symlink := filepath.Join(binDir, "go")
	if _, err := os.Lstat(symlink); err == nil {
		if err := os.Remove(symlink); err != nil {
			return fmt.Errorf("failed to remove existing symlink: %w", err)
		}
	}

	versionGoBin := filepath.Join(usr.HomeDir, "sdk", normalizeVersionName(version), "bin", "go")
	if err := os.Symlink(versionGoBin, symlink); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	fmt.Printf("Now using Go %s\n", version)
	fmt.Printf("Please add the following to your shell's config file:\n")
	fmt.Printf("export PATH=%s:$PATH\n", binDir)

	return nil
}

func removeVersion(version string) error {
	installed, err := isVersionInstalled(version)
	if err != nil {
		return err
	}

	if !installed {
		return fmt.Errorf("%w: %s", ErrVersionNotFound, version)
	}

	usr, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	versionDir := filepath.Join(usr.HomeDir, "sdk", normalizeVersionName(version))
	if err := os.RemoveAll(versionDir); err != nil {
		return fmt.Errorf("failed to remove Go version: %w", err)
	}

	// If the removed version is the active one, remove the symlink
	binDir := filepath.Join(usr.HomeDir, gvsDir, "bin")

	symlink := filepath.Join(binDir, "go")
	if _, err := os.Lstat(symlink); err == nil {
		link, err := os.Readlink(symlink)
		if err != nil {
			return fmt.Errorf("failed to read symlink: %w", err)
		}

		if link == filepath.Join(versionDir, "bin", "go") {
			if err := os.Remove(symlink); err != nil {
				return fmt.Errorf("failed to remove symlink: %w", err)
			}
		}
	}

	fmt.Printf("Go version %s removed successfully.\n", version)

	return nil
}

func isVersionInstalled(version string) (bool, error) {
	usr, err := user.Current()
	if err != nil {
		return false, fmt.Errorf("failed to get current user: %w", err)
	}

	versionDir := filepath.Join(usr.HomeDir, "sdk", normalizeVersionName(version))

	_, err = os.Stat(versionDir)
	if os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("failed to check if version is installed: %w", err)
	}

	return true, nil
}

func normalizeVersionName(version string) string {
	if !strings.HasPrefix(version, "go") {
		return "go" + version
	}

	return version
}

func printUsage() {
	fmt.Println("Usage: gvm <command> [arguments]")
	fmt.Println("Commands:")
	fmt.Println("  list-remote [-limit <n>] List all available Go versions (defaults to 10)")
	fmt.Println("  list                List all installed Go versions")
	fmt.Println("  install <version>   Install a new Go version")
	fmt.Println("  use <version>       Switch to a Go version")
	fmt.Println("  remove <version>    Uninstall a Go version")
	fmt.Println("  completion          Generate shell completion script")
}

func handleCompletion() {
	if len(os.Args) > 2 && os.Args[2] == "__complete" {
		complete()

		return
	}

	if len(os.Args) > 2 && os.Args[2] == "bash" {
		executable, err := os.Executable()
		if err != nil {
			// Fallback to just the command name if we can't get the full path
			executable = "gvs"
		}

		script := strings.ReplaceAll(bashCompletionScript, "@@GVM_EXECUTABLE@@", executable)
		fmt.Print(script)

		return
	}

	fmt.Println("Usage: gvm completion [bash]")
	os.Exit(1)
}

func complete() {
	currentWord := ""
	if len(os.Args) > 3 {
		currentWord = os.Args[3]
	}

	prevWord := ""
	if len(os.Args) > 4 {
		prevWord = os.Args[4]
	}

	switch prevWord {
	case "use", "remove":
		versions, err := getInstalledVersions()
		if err != nil {
			return
		}

		for _, v := range versions {
			if strings.HasPrefix(v, currentWord) {
				fmt.Println(v)
			}
		}
	case "gvs":
		subcommands := []string{"list", "list-remote", "install", "use", "remove", "completion"}

		for _, cmd := range subcommands {
			if strings.HasPrefix(cmd, currentWord) {
				fmt.Println(cmd)
			}
		}
	}
}

func getInstalledVersions() ([]string, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	versionsDir := filepath.Join(usr.HomeDir, "sdk")

	dirs, err := os.ReadDir(versionsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read versions directory: %w", err)
	}

	var versions []string

	for _, dir := range dirs {
		if dir.IsDir() {
			versions = append(versions, dir.Name())
		}
	}

	return versions, nil
}

const bashCompletionScript = `#compdef gvs
_gvm_completions() {
    local cur prev
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    local words
    # Call the Go binary with a hidden command to get completion words.
    # We pass the current and previous words for context.
    words=$(@@GVM_EXECUTABLE@@ completion __complete "${cur}" "${prev}")

    COMPREPLY=($(compgen -W "${words}" -- "${cur}"))
}

complete -F _gvm_completions gvs
`
