package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

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

func main() {
	// Check if node_modules directory exists
	if _, err := os.Stat("node_modules"); os.IsNotExist(err) {
		fmt.Println("node_modules directory not found.")
		return
	} else if err != nil {
		fmt.Println("Error checking directory:", err)
		return
	}

	// Calculate directory size
	size, err := calculateDirSize("node_modules")
	if err != nil {
		fmt.Println("Error calculating size:", err)
		return
	}

	// Display size in human readable format
	fmt.Printf("Total size of node_modules: %s\n", formatSize(size))

	// Ask user for deletion confirmation
	fmt.Print("Do you want to delete node_modules? (Y/N): ")
	var response string
	fmt.Scanln(&response)

	// If user chose "Y" or "y", delete the directory
	if response == "Y" || response == "y" {
		err := os.RemoveAll("node_modules")
		if err != nil {
			fmt.Println("Error deleting directory:", err)
		} else {
			fmt.Println("node_modules directory successfully deleted.")
		}
	} else {
		fmt.Println("node_modules directory was not deleted.")
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
