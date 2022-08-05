package main

import (
	"os"

	nodestart "github.com/paketo-buildpacks/node-start"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

func main() {
	nodeApplicationFinder := nodestart.NewNodeApplicationFinder()
	logger := scribe.NewEmitter(os.Stdout).WithLevel(os.Getenv("BP_LOG_LEVEL"))
	
	packit.Run(
		nodestart.Detect(nodeApplicationFinder),
		nodestart.Build(
			nodeApplicationFinder,
			logger,
		),
	)
}
