package nodestart

import (
	"fmt"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
)

func Build(applicationFinder ApplicationFinder, logger scribe.Logger) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		file, err := applicationFinder.Find(context.WorkingDir)
		if err != nil {
			return packit.BuildResult{}, err
		}

		command := fmt.Sprintf("node %s", file)

		logger.Process("Assigning launch processes")
		logger.Subprocess("web: %s", command)

		return packit.BuildResult{
			Launch: packit.LaunchMetadata{
				Processes: []packit.Process{
					{
						Type:    "web",
						Command: command,
					},
				},
			},
		}, nil
	}
}
