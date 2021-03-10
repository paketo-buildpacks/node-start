package main

import (
	"os"

	nodestart "github.com/paketo-buildpacks/node-start"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
)

func main() {
	logger := scribe.NewLogger(os.Stdout)
	nodeApplicationFinder := nodestart.NewNodeApplicationFinder()
	packit.Run(
		nodestart.Detect(nodeApplicationFinder),
		nodestart.Build(nodeApplicationFinder, logger),
	)
}
