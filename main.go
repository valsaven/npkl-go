package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"golang.org/x/term"
)

type NodeModulesDir struct {
	Path     string
	Size     int64
	Selected bool
}

// formatSize converts size in bytes to human readable format
func formatSize(bytes int64) string {
	const (
		B  = 1
		KB = 1024 * B
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

// findNodeModules recursively finds all node_modules directories
func findNodeModules(rootDir string) ([]NodeModulesDir, error) {
	var dirs []NodeModulesDir
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name() == "node_modules" {
			size, err := calculateDirSize(path)
			if err != nil {
				return err
			}
			dirs = append(dirs, NodeModulesDir{
				Path:     path,
				Size:     size,
				Selected: false,
			})
			return filepath.SkipDir
		}
		return nil
	})
	return dirs, err
}

// clearScreen clears the terminal screen
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

// displayDirs shows the list of directories with selection status
func displayDirs(dirs []NodeModulesDir, currentIdx int) {
	clearScreen()
	fmt.Println("Found node_modules directories:")
	fmt.Println("Use up/down arrows to navigate, space to select, Enter to confirm")
	fmt.Println()

	var totalSelectedSize int64
	for i, dir := range dirs {
		prefix := "[ ]"
		if dir.Selected {
			prefix = "[✓]"
			totalSelectedSize += dir.Size
		}
		cursor := "  "
		if i == currentIdx {
			cursor = "> "
		}
		fmt.Printf("%s%s %s (%s)\n", cursor, prefix, dir.Path, formatSize(dir.Size))
	}

	fmt.Println("\n----------------------------------------")
	fmt.Printf("Total size of selected directories: %s\n", formatSize(totalSelectedSize))
}

func main() {
	// Get current directory
	rootDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	// Find all node_modules directories
	dirs, err := findNodeModules(rootDir)
	if err != nil {
		fmt.Println("Error searching directories:", err)
		return
	}

	if len(dirs) == 0 {
		fmt.Println("No node_modules directories found.")
		return
	}

	// Switch terminal to raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("Error configuring terminal:", err)
		return
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Interactive selection
	currentIdx := 0
	displayDirs(dirs, currentIdx)

	buffer := make([]byte, 3)
	for {
		os.Stdin.Read(buffer)

		switch {
		case buffer[0] == 3: // Ctrl+C
			clearScreen()
			return
		case buffer[0] == 13: // Enter
			goto processSelection
		case buffer[0] == 32: // Space
			dirs[currentIdx].Selected = !dirs[currentIdx].Selected
		case buffer[0] == 27 && buffer[1] == 91: // Arrow keys
			switch buffer[2] {
			case 65: // Up
				if currentIdx > 0 {
					currentIdx--
				}
			case 66: // Down
				if currentIdx < len(dirs)-1 {
					currentIdx++
				}
			}
		}

		displayDirs(dirs, currentIdx)
	}

processSelection:
	clearScreen()
	selectedDirs := make([]NodeModulesDir, 0)
	for _, dir := range dirs {
		if dir.Selected {
			selectedDirs = append(selectedDirs, dir)
		}
	}

	if len(selectedDirs) == 0 {
		fmt.Println("Nothing selected for deletion.")
		return
	}

	fmt.Println("The following directories will be deleted:")
	var totalSize int64
	for _, dir := range selectedDirs {
		fmt.Printf("- %s (%s)\n", dir.Path, formatSize(dir.Size))
		totalSize += dir.Size
	}
	fmt.Printf("\nTotal size: %s\n", formatSize(totalSize))
	fmt.Print("\nConfirm deletion (Y/N): ")

	// Switch back to normal mode for confirmation
	term.Restore(int(os.Stdin.Fd()), oldState)

	var response string
	fmt.Scanln(&response)

	if strings.ToUpper(response) == "Y" {
		for _, dir := range selectedDirs {
			err := os.RemoveAll(dir.Path)
			if err != nil {
				fmt.Printf("Error deleting %s: %v\n", dir.Path, err)
			} else {
				fmt.Printf("Successfully deleted: %s\n", dir.Path)
			}
		}
	} else {
		fmt.Println("Operation cancelled.")
	}
}

// calculateDirSize recursively calculates the total size of all files in a directory
func calculateDirSize(dir string) (int64, error) {
	var totalSize int64
	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	// Directory walk function
	walkFunc := func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// If it's a directory, continue traversal
		if d.IsDir() {
			return nil
		}
		// For files, start a goroutine to calculate size
		wg.Add(1)
		go func() {
			defer wg.Done()
			fileInfo, err := d.Info()
			if err != nil {
				select {
				case errChan <- err:
				default:
				}
				return
			}
			// Atomically add file size to total sum
			atomic.AddInt64(&totalSize, fileInfo.Size())
		}()
		return nil
	}

	// Start recursive directory traversal
	err := filepath.WalkDir(dir, walkFunc)
	if err != nil {
		return 0, err
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Check for any errors
	if err := <-errChan; err != nil {
		return 0, err
	}

	return totalSize, nil
}
