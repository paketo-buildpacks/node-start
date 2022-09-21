package nodestart

import (
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/libreload-packit"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fs"
)

//go:generate faux --interface ApplicationFinder --output fakes/application_finder.go
type ApplicationFinder interface {
	Find(workingDir, launchpoint, projectPath string) (string, error)
}

type Reloader libreload.Reloader

//go:generate faux --interface Reloader --output fakes/reloader.go

func Detect(applicationFinder ApplicationFinder, reloader Reloader) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		_, err := applicationFinder.Find(context.WorkingDir, os.Getenv("BP_LAUNCHPOINT"), os.Getenv("BP_NODE_PROJECT_PATH"))
		if err != nil {
			return packit.DetectResult{}, err
		}

		requirements := []packit.BuildPlanRequirement{newLaunchRequirement("node")}

		if packageJsonExists, err := fs.Exists(filepath.Join(context.WorkingDir, os.Getenv("BP_NODE_PROJECT_PATH"), "package.json")); err != nil {
			return packit.DetectResult{}, err
		} else if packageJsonExists {
			requirements = append(requirements, newLaunchRequirement("node_modules"))
		}

		if shouldReload, err := reloader.ShouldEnableLiveReload(); err != nil {
			return packit.DetectResult{}, err
		} else if shouldReload {
			requirements = append(requirements, newLaunchRequirement("watchexec"))
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Requires: requirements,
			},
		}, nil
	}
}

func newLaunchRequirement(name string) packit.BuildPlanRequirement {
	return packit.BuildPlanRequirement{
		Name: name,
		Metadata: map[string]interface{}{
			"launch": true,
		},
	}
}
