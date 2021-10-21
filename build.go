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

		processes := []packit.Process{
			{
				Type:    "web",
				Command: command,
			},
		}

		shouldReload, err := checkLiveReloadEnabled()
		if err != nil {
			return packit.BuildResult{}, err
		}

		if shouldReload {
			processes = []packit.Process{
				{
					Type:    "web",
					Command: fmt.Sprintf(`watchexec --restart --watch /workspace "%s"`, command),
				},
				{
					Type:    "no-reload",
					Command: command,
				},
			}
		}

		logger.Process("Assigning launch processes")

		for _, process := range processes {
			logger.Subprocess("%s: %s", process.Type, process.Command)
		}

		return packit.BuildResult{
			Launch: packit.LaunchMetadata{
				Processes: processes,
			},
		}, nil
	}
}
