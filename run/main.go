package main

import (
	"os"

	nodestart "github.com/paketo-buildpacks/node-start"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
)

func main() {
	logger := scribe.NewLogger(os.Stdout)
	nodeApplicationDetector := nodestart.NewNodeApplicationDetector()
	packit.Run(
		nodestart.Detect(nodeApplicationDetector),
		nodestart.Build(nodeApplicationDetector, logger),
	)
}
