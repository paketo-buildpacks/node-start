package nodestart_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/libreload-packit"
	nodestart "github.com/paketo-buildpacks/node-start/v2"
	"github.com/paketo-buildpacks/node-start/v2/fakes"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir  string
		workingDir string
		cnbDir     string
		buffer     *bytes.Buffer

		reloader *fakes.Reloader

		buildContext packit.BuildContext
		build        packit.BuildFunc
	)

	it.Before(func() {
		layersDir = t.TempDir()
		cnbDir = t.TempDir()
		workingDir = t.TempDir()

		Expect(os.WriteFile(filepath.Join(workingDir, "server.js"), nil, 0600)).To(Succeed())

		reloader = &fakes.Reloader{}

		buffer = bytes.NewBuffer(nil)
		logger := scribe.NewEmitter(buffer)

		buildContext = packit.BuildContext{
			WorkingDir: workingDir,
			CNBPath:    cnbDir,
			Stack:      "some-stack",
			BuildpackInfo: packit.BuildpackInfo{
				Name:    "Some Buildpack",
				Version: "some-version",
			},
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{},
			},
			Layers: packit.Layers{Path: layersDir},
		}
		build = nodestart.Build(logger, reloader)
	})

	it("returns a result that provides a node start command", func() {
		result, err := build(buildContext)
		Expect(err).NotTo(HaveOccurred())

		Expect(result).To(Equal(packit.BuildResult{
			Plan: packit.BuildpackPlan{
				Entries: nil,
			},
			Layers: nil,
			Launch: packit.LaunchMetadata{
				Processes: []packit.Process{
					{
						Type:    "web",
						Command: "node",
						Args:    []string{"server.js"},
						Default: true,
						Direct:  true,
					},
				},
			},
		}))

		Expect(buffer.String()).To(ContainSubstring("Some Buildpack some-version"))
		Expect(buffer.String()).To(ContainSubstring("Assigning launch processes"))
		Expect(buffer.String()).To(ContainSubstring("node server.js"))
	})

	context("sniff test that files that end something other than js work", func() {
		it.Before(func() {
			layersDir = t.TempDir()
			cnbDir = t.TempDir()
			workingDir = t.TempDir()

			Expect(os.WriteFile(filepath.Join(workingDir, "app.mjs"), nil, 0600)).To(Succeed())

			reloader = &fakes.Reloader{}

			buffer = bytes.NewBuffer(nil)
			logger := scribe.NewEmitter(buffer)

			buildContext = packit.BuildContext{
				WorkingDir: workingDir,
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				BuildpackInfo: packit.BuildpackInfo{
					Name:    "Some Buildpack",
					Version: "some-version",
				},
				Plan: packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{},
				},
				Layers: packit.Layers{Path: layersDir},
			}
			build = nodestart.Build(logger, reloader)
		})

		it("returns a result that provides a node start command", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal(packit.BuildResult{
				Plan: packit.BuildpackPlan{
					Entries: nil,
				},
				Layers: nil,
				Launch: packit.LaunchMetadata{
					Processes: []packit.Process{
						{
							Type:    "web",
							Command: "node",
							Args:    []string{"app.mjs"},
							Default: true,
							Direct:  true,
						},
					},
				},
			}))

			Expect(buffer.String()).To(ContainSubstring("Some Buildpack some-version"))
			Expect(buffer.String()).To(ContainSubstring("Assigning launch processes"))
			Expect(buffer.String()).To(ContainSubstring("node app.mjs"))
		})
	})

	context("when live reload is enabled", func() {
		it.Before(func() {
			reloader.ShouldEnableLiveReloadCall.Returns.Bool = true
			reloader.TransformReloadableProcessesCall.Returns.Reloadable = packit.Process{
				Type:    "Reloadable",
				Command: "Reloadable-Command",
			}
			reloader.TransformReloadableProcessesCall.Returns.NonReloadable = packit.Process{
				Type:    "Nonreloadable",
				Command: "Nonreloadable-Command",
			}
		})

		it("adds a reloadable start command and makes it the default", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Launch.Processes).To(Equal([]packit.Process{
				{
					Type:    "web",
					Command: "Reloadable-Command",
				},
				{
					Type:    "no-reload",
					Command: "Nonreloadable-Command",
				},
			}))

			Expect(reloader.TransformReloadableProcessesCall.Receives.OriginalProcess).To(Equal(packit.Process{
				Type:    "web",
				Command: "node",
				Args:    []string{"server.js"},
				Default: true,
				Direct:  true,
			}))

			Expect(reloader.TransformReloadableProcessesCall.Receives.Spec).To(Equal(libreload.ReloadableProcessSpec{
				WatchPaths: []string{workingDir},
			}))

			Expect(buffer.String()).To(ContainSubstring("Some Buildpack some-version"))
			Expect(buffer.String()).To(ContainSubstring("Assigning launch processes"))
		})
	})

	context("failure cases", func() {
		context("when the application finding fails", func() {
			it.Before(func() {
				Expect(os.Remove(filepath.Join(workingDir, "server.js"))).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError(fmt.Errorf("could not find app in %s: expected one of server.js | server.cjs | server.mjs | app.js | app.cjs | app.mjs | main.js | main.cjs | main.mjs | index.js | index.cjs | index.mjs", workingDir)))
			})
		})

		context("when the reloader returns an error", func() {
			it.Before(func() {
				reloader.ShouldEnableLiveReloadCall.Returns.Error = errors.New("reloader error")
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError("reloader error"))
			})
		})
	})
}
