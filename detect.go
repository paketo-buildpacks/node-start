package nodestart

import (
	"github.com/paketo-buildpacks/packit"
)

//go:generate faux --interface ApplicationDetector --output fakes/application_detector.go
type ApplicationDetector interface {
	Detect(workingDir string) (string, error)
}

func Detect(applicationDetector ApplicationDetector) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		_, err := applicationDetector.Detect(context.WorkingDir)
		if err != nil {
			return packit.DetectResult{}, err
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"launch": true,
						},
					},
				},
			},
		}, nil
	}
}
