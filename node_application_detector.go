package nodestart

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit"
)

type NodeApplicationDetector struct{}

func NewNodeApplicationDetector() NodeApplicationDetector {
	return NodeApplicationDetector{}
}

func (n NodeApplicationDetector) Detect(workingDir string) (string, error) {
	files := []string{"server.js", "app.js", "main.js", "index.js"}
	for _, file := range files {
		_, err := os.Stat(filepath.Join(workingDir, file))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			} else {
				return "", err
			}
		}
		return file, nil
	}

	return "", packit.Fail.WithMessage("expected one of the following files to be in your application root: %s", strings.Join(files, " | "))
}
