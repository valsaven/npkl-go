package ui

import (
	"fmt"
	"npkl-go/pkg/directory"
	"npkl-go/pkg/utils"
)

// ClearScreen clears the terminal screen
func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

// DisplayDirs shows the list of directories with selection status
func DisplayDirs(dirs []directory.NodeModulesDir, currentIdx int) {
	ClearScreen()
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
		fmt.Printf("%s%s %s (%s)\n", cursor, prefix, dir.Path, utils.FormatSize(dir.Size))
	}

	fmt.Println("\n----------------------------------------")
	fmt.Printf("Total size of selected directories: %s\n", utils.FormatSize(totalSelectedSize))
}
