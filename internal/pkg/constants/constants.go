package constants

import (
	"fmt"
	"os"
	"path/filepath"
)

func FindProjectRoot() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if fileExists(filepath.Join(currentDir, "go.mod")) {
			return currentDir, nil
		}

		parentDir := filepath.Dir(currentDir)
		if currentDir == parentDir {
			break
		}
		currentDir = parentDir
	}

	return "", fmt.Errorf("project root not found")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
