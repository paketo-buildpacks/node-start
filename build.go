package nodestart

import (
	"os"

	"github.com/paketo-buildpacks/libreload-packit"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

func Build(applicationFinder ApplicationFinder, logger scribe.Emitter, reloader Reloader) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		file, err := applicationFinder.Find(context.WorkingDir, os.Getenv("BP_LAUNCHPOINT"), os.Getenv("BP_NODE_PROJECT_PATH"))
		if err != nil {
			return packit.BuildResult{}, err
		}

		originalProcess := packit.Process{
			Type:    "web",
			Command: "node",
			Args:    []string{file},
			Default: true,
			Direct:  true,
		}

		var processes []packit.Process
		if shouldEnableReload, err := reloader.ShouldEnableLiveReload(); err != nil {
			return packit.BuildResult{}, err
		} else if shouldEnableReload {
			nonReloadableProcess, reloadableProcess := reloader.TransformReloadableProcesses(originalProcess, libreload.ReloadableProcessSpec{
				WatchPaths: []string{context.WorkingDir},
			})
			reloadableProcess.Type = "web"
			nonReloadableProcess.Type = "no-reload"
			processes = append(processes, reloadableProcess, nonReloadableProcess)
		} else {
			processes = append(processes, originalProcess)
		}

		logger.LaunchProcesses(processes)

		return packit.BuildResult{
			Launch: packit.LaunchMetadata{
				Processes: processes,
			},
		}, nil
	}
}
