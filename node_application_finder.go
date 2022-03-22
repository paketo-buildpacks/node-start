package nodestart

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit/v2"
)

type NodeApplicationFinder struct{}

func NewNodeApplicationFinder() NodeApplicationFinder {
	return NodeApplicationFinder{}
}

func (n NodeApplicationFinder) Find(workingDir, launchpoint, projectPath string) (string, error) {
	if launchpoint != "" {
		if _, err := os.Stat(filepath.Join(workingDir, launchpoint)); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return "", fmt.Errorf("expected value derived from BP_LAUNCHPOINT [%s] to be an existing file", launchpoint)
			}

			return "", err
		}

		return filepath.Clean(launchpoint), nil
	}

	files := []string{"server.js", "app.js", "main.js", "index.js"}
	for _, file := range files {
		_, err := os.Stat(filepath.Join(workingDir, projectPath, file))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}

			return "", err
		}

		return filepath.Join(projectPath, file), nil
	}

	return "", packit.Fail.WithMessage("could not find app in %s: expected one of %s", filepath.Clean(filepath.Join(workingDir, projectPath)), strings.Join(files, " | "))
}
