package nodestart

import (
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
)

func Build(logger scribe.Logger) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		command := "node server.js"

		logger.Process("Assigning launch processes")
		logger.Subprocess("web: %s", command)

		return packit.BuildResult{
			Processes: []packit.Process{
				{
					Type:    "web",
					Command: command,
				},
			},
		}, nil
	}
}
