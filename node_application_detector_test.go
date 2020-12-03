package nodestart_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	nodestart "github.com/paketo-buildpacks/node-start"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testNodeApplicationDetector(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir              string
		nodeApplicationDetector nodestart.NodeApplicationDetector
	)

	it.Before(func() {
		var err error
		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		nodeApplicationDetector = nodestart.NewNodeApplicationDetector()
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("when there is at least one expected application file in the working dir", func() {
		it.Before(func() {
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "server.js"), nil, 0644)).To(Succeed())
		})

		it("detects", func() {
			file, err := nodeApplicationDetector.Detect(workingDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(file).To(Equal("server.js"))
		})
	})

	context("when there is more than one expected application file in the working dir", func() {
		it.Before(func() {
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "app.js"), nil, 0644)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "main.js"), nil, 0644)).To(Succeed())
		})

		it("returns the highest priority file", func() {
			file, err := nodeApplicationDetector.Detect(workingDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(file).To(Equal("app.js"))
		})
	})

	context("failure cases", func() {
		context("when the working dir cannot be stated", func() {
			it.Before(func() {
				Expect(os.Chmod(workingDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
			})

			it("fails with helpful error", func() {
				_, err := nodeApplicationDetector.Detect(workingDir)
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when there are no expected files in the working dir", func() {
			it("fails with helpful error", func() {
				_, err := nodeApplicationDetector.Detect(workingDir)
				Expect(err).To(MatchError(ContainSubstring("expected one of the following files to be in your application root:")))
			})
		})
	})
}
