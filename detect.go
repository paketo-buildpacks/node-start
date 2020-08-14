package nodestart

import (
	"fmt"
	"github.com/paketo-buildpacks/packit"
	"path/filepath"
)

func Detect() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		files, err := filepath.Glob(filepath.Join(context.WorkingDir, "*.js"))
		if err != nil {
			return packit.DetectResult{}, fmt.Errorf("file glob function failed: %w", err)
		}
		if len(files) == 0 {
			return packit.DetectResult{}, packit.Fail
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
