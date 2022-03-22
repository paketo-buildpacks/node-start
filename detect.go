package nodestart

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fs"
)

//go:generate faux --interface ApplicationFinder --output fakes/application_finder.go
type ApplicationFinder interface {
	Find(workingDir, launchpoint, projectPath string) (string, error)
}

func Detect(applicationFinder ApplicationFinder) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		_, err := applicationFinder.Find(context.WorkingDir, os.Getenv("BP_LAUNCHPOINT"), os.Getenv("BP_NODE_PROJECT_PATH"))
		if err != nil {
			return packit.DetectResult{}, err
		}

		requirements := []packit.BuildPlanRequirement{
			{
				Name: "node",
				Metadata: map[string]interface{}{
					"launch": true,
				},
			},
		}

		exists, err := fs.Exists(filepath.Join(context.WorkingDir, os.Getenv("BP_NODE_PROJECT_PATH"), "package.json"))
		if err != nil {
			return packit.DetectResult{}, err
		}

		if exists {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "node_modules",
				Metadata: map[string]interface{}{
					"launch": true,
				},
			})
		}

		shouldReload, err := checkLiveReloadEnabled()
		if err != nil {
			return packit.DetectResult{}, err
		}

		if shouldReload {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: "watchexec",
				Metadata: map[string]interface{}{
					"launch": true,
				},
			})
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{},
				Requires: requirements,
			},
		}, nil
	}
}

func checkLiveReloadEnabled() (bool, error) {
	if reload, ok := os.LookupEnv("BP_LIVE_RELOAD_ENABLED"); ok {
		shouldEnableReload, err := strconv.ParseBool(reload)
		if err != nil {
			return false, fmt.Errorf("failed to parse BP_LIVE_RELOAD_ENABLED value %s: %w", reload, err)
		}
		return shouldEnableReload, nil
	}
	return false, nil
}
