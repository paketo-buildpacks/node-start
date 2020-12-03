package nodestart_test

import (
	"errors"
	"testing"

	nodestart "github.com/paketo-buildpacks/node-start"
	"github.com/paketo-buildpacks/node-start/fakes"
	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		applicationDetector *fakes.ApplicationDetector

		detect packit.DetectFunc
	)

	it.Before(func() {
		applicationDetector = &fakes.ApplicationDetector{}

		detect = nodestart.Detect(applicationDetector)
	})

	context("when an application is detected in the working dir", func() {
		it("detects", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: "working-dir",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"launch": true,
						},
					},
				},
			}))

			Expect(applicationDetector.DetectCall.Receives.WorkingDir).To(Equal("working-dir"))
		})
	})

	context("failure cases", func() {
		context("when the application detector fails", func() {
			it.Before(func() {
				applicationDetector.DetectCall.Returns.Error = errors.New("detector failed")
			})

			it("fails with helpful error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: "working-dir",
				})
				Expect(err).To(MatchError("detector failed"))
			})
		})
	})
}
