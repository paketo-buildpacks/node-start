package nodestart_test

import (
	"fmt"
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

	context("when BP_LAUNCHPOINT is set", func() {
		it.Before(func() {
			Expect(os.Mkdir(filepath.Join(workingDir, "src"), os.ModePerm)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "src", "launchpoint.js"), nil, 0644)).To(Succeed())
			Expect(os.Setenv("BP_LAUNCHPOINT", "./src/launchpoint.js")).To(Succeed())
		})

		it.After(func() {
			os.Unsetenv("BP_LAUNCHPOINT")
		})

		it("returns the highest priority file", func() {
			file, err := nodeApplicationDetector.Detect(workingDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(file).To(Equal(filepath.Join("src", "launchpoint.js")))
		})

	}, spec.Sequential())

	context("failure cases", func() {
		context("when the file specified by BP_LAUNCHPOINT does not exist", func() {
			it.Before(func() {
				Expect(os.Mkdir(filepath.Join(workingDir, "src"), os.ModePerm)).To(Succeed())
				Expect(os.Setenv("BP_LAUNCHPOINT", "./src/launchpoint.js")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("BP_LAUNCHPOINT")).To(Succeed())
			})

			it("fails with helpful error", func() {
				_, err := nodeApplicationDetector.Detect(workingDir)
				Expect(err).To(
					MatchError(
						ContainSubstring(
							fmt.Sprintf("expected value derived from BP_LAUNCHPOINT [%s] to be an existing file", filepath.Join(workingDir, "src", "launchpoint.js")),
						),
					),
				)
			})
		}, spec.Sequential())

		context("when os.Stat cannot be performed on the launchpoint dir", func() {
			it.Before(func() {
				Expect(os.Mkdir(filepath.Join(workingDir, "src"), os.ModePerm)).To(Succeed())
				Expect(os.Chmod(filepath.Join(workingDir, "src"), 0000)).To(Succeed())
				Expect(os.Setenv("BP_LAUNCHPOINT", "./src/launchpoint.js")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(filepath.Join(workingDir, "src"), os.ModePerm)).To(Succeed())
				Expect(os.Unsetenv("BP_LAUNCHPOINT")).To(Succeed())
			})

			it("fails with helpful error", func() {
				_, err := nodeApplicationDetector.Detect(filepath.Join(workingDir))
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		}, spec.Sequential())

		context("when os.Stat() cannot be performed on the working dir", func() {
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
		}, spec.Sequential())

		context("when there are no expected files in the working dir", func() {
			it("fails with helpful error", func() {
				_, err := nodeApplicationDetector.Detect(workingDir)
				Expect(err).To(MatchError(ContainSubstring("expected one of the following files to be in your application root:")))
			})
		})
	})
}
