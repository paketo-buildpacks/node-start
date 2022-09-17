package nodestart_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	nodestart "github.com/paketo-buildpacks/node-start"
	"github.com/paketo-buildpacks/node-start/fakes"
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

		applicationFinder *fakes.ApplicationFinder

		buildContext packit.BuildContext
		build        packit.BuildFunc
	)

	it.Before(func() {
		var err error
		layersDir, err = os.MkdirTemp("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		applicationFinder = &fakes.ApplicationFinder{}
		applicationFinder.FindCall.Returns.String = "server.js"

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
		build = nodestart.Build(applicationFinder, logger)
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
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

		Expect(applicationFinder.FindCall.Receives.WorkingDir).To(Equal(workingDir))

		Expect(buffer.String()).To(ContainSubstring("Some Buildpack some-version"))
		Expect(buffer.String()).To(ContainSubstring("Assigning launch processes"))
		Expect(buffer.String()).To(ContainSubstring("node server.js"))
	})

	context("when BP_LIVE_RELOAD_ENABLED=true in the build environment", func() {
		it.Before(func() {
			os.Setenv("BP_LIVE_RELOAD_ENABLED", "true")
		})

		it.After(func() {
			os.Unsetenv("BP_LIVE_RELOAD_ENABLED")
		})

		it("adds a reloadable start command and makes it the default", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Launch.Processes).To(Equal([]packit.Process{
				{
					Type:    "web",
					Command: "watchexec",
					Args: []string{
						"--restart",
						"--watch", workingDir,
						"--shell", "none",
						"--",
						"node", "server.js",
					},
					Default: true,
					Direct:  true,
				},
				{
					Type:    "no-reload",
					Command: "node",
					Args:    []string{"server.js"},
					Direct:  true,
				},
			}))

			Expect(buffer.String()).To(ContainSubstring("Some Buildpack some-version"))
			Expect(buffer.String()).To(ContainSubstring("Assigning launch processes"))
			Expect(buffer.String()).To(ContainSubstring(fmt.Sprintf(`watchexec --restart --watch %s --shell none -- node server.js`, workingDir)))
			Expect(buffer.String()).To(ContainSubstring("node server.js"))
		})
	})

	context("failure cases", func() {
		context("when the application finding fails", func() {
			it.Before(func() {
				applicationFinder.FindCall.Returns.Error = errors.New("failed to find application")
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError("failed to find application"))
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
				_, err := build(buildContext)
				Expect(err).To(MatchError(ContainSubstring("failed to parse BP_LIVE_RELOAD_ENABLED value not-a-bool")))
			})
		})
	})
}
