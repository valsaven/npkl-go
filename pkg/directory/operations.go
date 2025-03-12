package directory

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

// calculateFileSize calculates the size of a single file and adds it to the total sum
func calculateFileSize(d os.DirEntry, totalSize *int64, errChan chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	fileInfo, err := d.Info()
	if err != nil {
		select {
		case errChan <- err:
		default:
		}
		return
	}
	atomic.AddInt64(totalSize, fileInfo.Size())
}

// FindNodeModules recursively finds all node_modules directories
func FindNodeModules(rootDir string) ([]NodeModulesDir, error) {
	var dirs []NodeModulesDir
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name() == "node_modules" {
			size, err := CalculateDirSize(path)
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

// CalculateDirSize recursively calculates the total size of all files in a directory
func CalculateDirSize(dir string) (int64, error) {
	var totalSize int64
	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	walkFunc := func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		wg.Add(1)
		go calculateFileSize(d, &totalSize, errChan, &wg)
		return nil
	}

	err := filepath.WalkDir(dir, walkFunc)
	if err != nil {
		return 0, err
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	if err := <-errChan; err != nil {
		return 0, err
	}

	return totalSize, nil
}
