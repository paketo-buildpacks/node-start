package nodestart

import (
	"errors"
	"os"
	"path/filepath"
)

type NodeApplicationFinder struct{}

func NewNodeApplicationFinder() NodeApplicationFinder {
	return NodeApplicationFinder{}
}

func (n NodeApplicationFinder) Find(workingDir string) (string, error) {

	launchpoint := os.Getenv("BP_LAUNCHPOINT")
	if launchpoint != "" {
		if _, err := os.Stat(filepath.Join(workingDir, launchpoint)); err != nil {
			launchErr := launchpointError(launchpoint)
			if errors.Is(err, os.ErrNotExist) {
				return "", launchErr
			} else {
				return "", err
			}
		}
		return filepath.Clean(launchpoint), nil
	}

	projectPath := os.Getenv("BP_NODE_PROJECT_PATH")

	files := []string{"server.js", "app.js", "main.js", "index.js"}
	for _, file := range files {
		_, err := os.Stat(filepath.Join(workingDir, projectPath, file))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			} else {
				return "", err
			}
		}
		return filepath.Join(projectPath, file), nil
	}
	targetError := targetFileError{expectedFiles: files, projectPath: filepath.Clean(projectPath)}
	return "", targetError
}
