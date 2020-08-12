package main

import (
	nodestart "github.com/paketo-buildpacks/node-start"
	"github.com/paketo-buildpacks/packit"
)

func main() {
	packit.Run(nodestart.Detect(), nodestart.Build())
}
