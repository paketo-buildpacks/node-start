package main

import (
	"os"

	nodestart "github.com/paketo-buildpacks/node-start"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

func main() {
	logger := scribe.NewLogger(os.Stdout)
	nodeApplicationFinder := nodestart.NewNodeApplicationFinder()
	packit.Run(
		nodestart.Detect(nodeApplicationFinder),
		nodestart.Build(nodeApplicationFinder, logger),
	)
}
