package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"

	"npkl-go/pkg/directory"
	"npkl-go/pkg/ui"
	"npkl-go/pkg/utils"
)

func main() {
	rootDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	dirs, err := directory.FindNodeModules(rootDir)
	if err != nil {
		fmt.Println("Error searching directories:", err)
		return
	}

	if len(dirs) == 0 {
		fmt.Println("No node_modules directories found.")
		return
	}

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("Error configuring terminal:", err)
		return
	}
	if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
		fmt.Println("Error restoring terminal state:", err)
		return
	}
	defer func() {
		if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
			fmt.Println("Error restoring terminal state:", err)
		}
	}()

	currentIdx := 0
	ui.DisplayDirs(dirs, currentIdx)

	buffer := make([]byte, 3)
	for {
		if _, err := os.Stdin.Read(buffer); err != nil {
			fmt.Println("Error reading from stdin:", err)
			return
		}

		switch {
		case buffer[0] == 3: // Ctrl+C
			ui.ClearScreen()
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

		ui.DisplayDirs(dirs, currentIdx)
	}

processSelection:
	ui.ClearScreen()
	selectedDirs := make([]directory.NodeModulesDir, 0)
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
		fmt.Printf("- %s (%s)\n", dir.Path, utils.FormatSize(dir.Size))
		totalSize += dir.Size
	}
	fmt.Printf("\nTotal size: %s\n", utils.FormatSize(totalSize))
	fmt.Print("\nConfirm deletion (Y/N): ")

	if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
		fmt.Println("Error restoring terminal state:", err)
		return
	}

	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		fmt.Println("Error reading user input:", err)
		return
	}

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
