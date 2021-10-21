package nodestart_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
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

		applicationFinder *fakes.ApplicationFinder

		detect     packit.DetectFunc
		workingDir string
	)

	it.Before(func() {
		var err error
		applicationFinder = &fakes.ApplicationFinder{}
		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		applicationFinder.FindCall.Returns.String = "server.js"

		detect = nodestart.Detect(applicationFinder)
	})

	context("when an application is detected in the working dir", func() {
		it.Before(func() {
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "server.js"), []byte(nil), 0644)).To(Succeed())
		})

		it("detects", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
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

			Expect(applicationFinder.FindCall.Receives.WorkingDir).To(Equal(workingDir))
		})

		context("and BP_LIVE_RELOAD_ENABLED=true in the build environment", func() {
			it.Before(func() {
				os.Setenv("BP_LIVE_RELOAD_ENABLED", "true")
			})

			it.After(func() {
				os.Unsetenv("BP_LIVE_RELOAD_ENABLED")
			})

			it("requires watchexec at launch time", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan.Requires).To(Equal([]packit.BuildPlanRequirement{
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"launch": true,
						},
					},
					{
						Name: "watchexec",
						Metadata: map[string]interface{}{
							"launch": true,
						},
					},
				},
				))
			})
		})
	})

	context("when BP_LAUNCHPOINT file does not exist", func() {
		it.Before(func() {
			applicationFinder.FindCall.Returns.Error = nodestart.NewLaunchpointError("launchpoint")
		})
		it("fails detection", func() {
			_, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).To(Equal(packit.Fail.WithMessage("expected value derived from BP_LAUNCHPOINT [launchpoint] to be an existing file")))
			Expect(applicationFinder.FindCall.Receives.WorkingDir).To(Equal(workingDir))
		})
	})

	context("when no application is detected in the working dir", func() {
		it.Before(func() {
			applicationFinder.FindCall.Returns.Error = nodestart.NewTargetFileError([]string{"someFile"}, "somePath")
		})
		it("fails detection", func() {
			_, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).To(Equal(packit.Fail.WithMessage("expected one of the following files to be in your application root (somePath): someFile")))
			Expect(applicationFinder.FindCall.Receives.WorkingDir).To(Equal(workingDir))
		})
	})

	context("failure cases", func() {
		context("when the application finder fails", func() {
			it.Before(func() {
				applicationFinder.FindCall.Returns.Error = errors.New("finder failed")
			})

			it("fails with helpful error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError("finder failed"))
			})
		})

		context("when BP_LIVE_RELOAD_ENABLED is set to an invalid value", func() {
			it.Before(func() {
				os.Setenv("BP_LIVE_RELOAD_ENABLED", "not-a-bool")
			})

			it.After(func() {
				os.Unsetenv("BP_LIVE_RELOAD_ENABLED")
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError(ContainSubstring("failed to parse BP_LIVE_RELOAD_ENABLED value not-a-bool")))
			})
		})
	})
}
