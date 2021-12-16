package nodestart

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/paketo-buildpacks/packit/v2"
)

//go:generate faux --interface ApplicationFinder --output fakes/application_finder.go
type ApplicationFinder interface {
	Find(workingDir string) (string, error)
}

type launchpointError string

func NewLaunchpointError(launchpoint string) launchpointError {
	return launchpointError(launchpoint)
}
func (s launchpointError) Error() string {
	return fmt.Sprintf("expected value derived from BP_LAUNCHPOINT [%s] to be an existing file", string(s))
}

type targetFileError struct {
	expectedFiles []string
	projectPath   string
}

func NewTargetFileError(expectedFiles []string, projectPath string) targetFileError {
	return targetFileError{
		expectedFiles: expectedFiles,
		projectPath:   projectPath,
	}
}

func (t targetFileError) Error() string {
	return fmt.Sprintf("expected one of the following files to be in your application root (%s): %s", t.projectPath, strings.Join(t.expectedFiles, " | "))
}

func Detect(applicationFinder ApplicationFinder) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		_, err := applicationFinder.Find(context.WorkingDir)

		if _, ok := err.(launchpointError); ok {
			return packit.DetectResult{}, packit.Fail.WithMessage(err.Error())
		}

		if _, ok := err.(targetFileError); ok {
			return packit.DetectResult{}, packit.Fail.WithMessage(err.Error())
		}

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
