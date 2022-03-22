package nodestart

import (
	"os"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

func Build(applicationFinder ApplicationFinder, logger scribe.Emitter) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		file, err := applicationFinder.Find(context.WorkingDir, os.Getenv("BP_LAUNCHPOINT"), os.Getenv("BP_NODE_PROJECT_PATH"))
		if err != nil {
			return packit.BuildResult{}, err
		}

		command := "node"
		args := []string{file}

		processes := []packit.Process{
			{
				Type:    "web",
				Command: command,
				Args:    args,
				Default: true,
				Direct:  true,
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
					Command: "watchexec",
					Args: append([]string{
						"--restart",
						"--watch", context.WorkingDir,
						"--shell", "none",
						"--",
						command,
					}, args...),
					Direct:  true,
					Default: true,
				},
				{
					Type:    "no-reload",
					Command: command,
					Args:    args,
					Direct:  true,
				},
			}
		}

		logger.LaunchProcesses(processes)

		return packit.BuildResult{
			Launch: packit.LaunchMetadata{
				Processes: processes,
			},
		}, nil
	}
}
