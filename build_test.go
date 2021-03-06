package nodestart_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"testing"

	nodestart "github.com/paketo-buildpacks/node-start"
	"github.com/paketo-buildpacks/node-start/fakes"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
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

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		layersDir, err = ioutil.TempDir("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = ioutil.TempDir("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		applicationFinder = &fakes.ApplicationFinder{}
		applicationFinder.FindCall.Returns.String = "server.js"

		buffer = bytes.NewBuffer(nil)
		logger := scribe.NewLogger(buffer)
		build = nodestart.Build(applicationFinder, logger)
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	it("returns a result that provides a node start command", func() {
		result, err := build(packit.BuildContext{
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
		})
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
						Command: "node server.js",
					},
				},
			},
		}))

		Expect(applicationFinder.FindCall.Receives.WorkingDir).To(Equal(workingDir))

		Expect(buffer.String()).To(ContainSubstring("Some Buildpack some-version"))
		Expect(buffer.String()).To(ContainSubstring("Assigning launch processes"))
		Expect(buffer.String()).To(ContainSubstring("node server.js"))
	})

	context("failure cases", func() {
		context("when the application finding fails", func() {
			it.Before(func() {
				applicationFinder.FindCall.Returns.Error = errors.New("failed to find application")

			})
			it("returns an error", func() {
				_, err := build(packit.BuildContext{
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
				})
				Expect(err).To(MatchError("failed to find application"))
			})
		})
	})
}
