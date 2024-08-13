package nodestart_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	nodestart "github.com/paketo-buildpacks/node-start/v2"
	"github.com/paketo-buildpacks/node-start/v2/fakes"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		reloader *fakes.Reloader

		detect     packit.DetectFunc
		workingDir string
	)

	it.Before(func() {
		workingDir = t.TempDir()

		reloader = &fakes.Reloader{}

		detect = nodestart.Detect(reloader)
	})

	context("when an application is detected in the working dir", func() {
		it.Before(func() {
			t.Setenv("BP_NODE_PROJECT_PATH", "./src")
			t.Setenv("BP_LAUNCHPOINT", "./src/server.js")
			Expect(os.MkdirAll(filepath.Join(workingDir, "src"), os.ModePerm)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(workingDir, "src", "server.js"), nil, 0600)).To(Succeed())
		})

		it("detects", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"launch": true,
						},
					},
				},
			}))
		})

		context("when live reload is enabled", func() {
			it.Before(func() {
				reloader.ShouldEnableLiveReloadCall.Returns.Bool = true
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
			t.Setenv("BP_NODE_PROJECT_PATH", "./src")
			Expect(os.MkdirAll(filepath.Join(workingDir, "src"), os.ModePerm)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(workingDir, "src", "package.json"), []byte(`{}`), 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(workingDir, "src", "server.js"), nil, 0600)).To(Succeed())
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
			it("fails with helpful error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError(fmt.Errorf("could not find app in %s: expected one of server.js | app.js | main.js | index.js", workingDir)))
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

		context("when the reloader returns an error", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(workingDir, "server.js"), nil, 0600)).To(Succeed())
				reloader.ShouldEnableLiveReloadCall.Returns.Error = errors.New("reloader error")
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError("reloader error"))
			})
		}, spec.Sequential())
	})
}
