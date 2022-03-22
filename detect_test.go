package nodestart_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	nodestart "github.com/paketo-buildpacks/node-start"
	"github.com/paketo-buildpacks/node-start/fakes"
	"github.com/paketo-buildpacks/packit/v2"
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
		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		applicationFinder = &fakes.ApplicationFinder{}
		applicationFinder.FindCall.Returns.String = "./src/server.js"

		detect = nodestart.Detect(applicationFinder)
	})

	context("when an application is detected in the working dir", func() {
		it.Before(func() {
			os.Setenv("BP_NODE_PROJECT_PATH", "./src")
			os.Setenv("BP_LAUNCHPOINT", "./src/server.js")
			Expect(os.MkdirAll(filepath.Join(workingDir, "src"), os.ModePerm)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(workingDir, "src", "server.js"), nil, 0600)).To(Succeed())
		})

		it.After(func() {
			os.Unsetenv("BP_NODE_PROJECT_PATH")
			os.Unsetenv("BP_LAUNCHPOINT")
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
			Expect(applicationFinder.FindCall.Receives.Launchpoint).To(Equal("./src/server.js"))
			Expect(applicationFinder.FindCall.Receives.ProjectPath).To(Equal("./src"))
		})

		context("when BP_LIVE_RELOAD_ENABLED=true", func() {
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
				}))
			})
		})
	}, spec.Sequential())

	context("when a package.json is detected in the working dir", func() {
		it.Before(func() {
			os.Setenv("BP_NODE_PROJECT_PATH", "./src")
			Expect(os.MkdirAll(filepath.Join(workingDir, "src"), os.ModePerm)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(workingDir, "src", "package.json"), nil, 0600)).To(Succeed())
		})

		it.After(func() {
			os.Unsetenv("BP_NODE_PROJECT_PATH")
		})

		it("requires node_modules", func() {
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
					Name: "node_modules",
					Metadata: map[string]interface{}{
						"launch": true,
					},
				},
			}))
		})
	}, spec.Sequential())

	context("failure cases", func() {
		context("when the application finder fails", func() {
			it.Before(func() {
				applicationFinder.FindCall.Returns.Error = errors.New("application finder failed")
			})

			it("fails with helpful error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError("application finder failed"))
			})
		})

		context("when the package.json cannot be found", func() {
			it.Before(func() {
				Expect(os.Chmod(workingDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
			})

			it("fails with helpful error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
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
		}, spec.Sequential())
	})
}
