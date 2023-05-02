package nodestart

import (
	"os"

	"github.com/paketo-buildpacks/libnodejs"
	"github.com/paketo-buildpacks/libreload-packit"
	"github.com/paketo-buildpacks/packit/v2"
)

//go:generate faux --interface ApplicationFinder --output fakes/application_finder.go
type ApplicationFinder interface {
	Find(workingDir, launchpoint, projectPath string) (string, error)
}

type Reloader libreload.Reloader

//go:generate faux --interface Reloader --output fakes/reloader.go

func Detect(reloader Reloader) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		_, err := libnodejs.FindNodeApplication(context.WorkingDir)
		if err != nil {
			return packit.DetectResult{}, err
		}

		requirements := []packit.BuildPlanRequirement{newLaunchRequirement("node")}

		projectPath, err := libnodejs.FindProjectPath(context.WorkingDir)
		if err != nil {
			return packit.DetectResult{}, err
		}

		if _, err := libnodejs.ParsePackageJSON(projectPath); err != nil {
			if !os.IsNotExist(err) {
				return packit.DetectResult{}, err
			}
		} else {
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
