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

func testNodeApplicationFinder(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir            string
		nodeApplicationFinder nodestart.NodeApplicationFinder
	)

	it.Before(func() {
		var err error
		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		nodeApplicationFinder = nodestart.NewNodeApplicationFinder()
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("when BP_LAUNCHPOINT is set", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_LAUNCHPOINT", "./src/launchpoint.js")).To(Succeed())
		})

		it.After(func() {
			os.Unsetenv("BP_LAUNCHPOINT")
		})

		context("file specified by LAUNCHPOINT exists", func() {
			it.Before(func() {
				Expect(os.Mkdir(filepath.Join(workingDir, "src"), os.ModePerm)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(workingDir, "src", "launchpoint.js"), nil, 0644)).To(Succeed())

			})

			it.After(func() {
				Expect(os.RemoveAll(filepath.Join(workingDir, "src"))).To(Succeed())
			})

			it("returns the highest priority file", func() {
				file, err := nodeApplicationFinder.Find(workingDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(file).To(Equal(filepath.Join("src", "launchpoint.js")))
			})
		})

		context("file specified by LAUNCHPOINT DOES NOT exist", func() {
			it("returns the empty string and no error", func() {
				file, err := nodeApplicationFinder.Find(workingDir)
				Expect(err).To(MatchError(ContainSubstring("expected value derived from BP_LAUNCHPOINT [./src/launchpoint.js] to be an existing file")))
				Expect(file).To(Equal(""))
			})
		})
	}, spec.Sequential())

	context("BP_LAUNCHPOINT is NOT set && BP_NODE_PROJECT_PATH is set", func() {
		it.Before(func() {
			Expect(os.Mkdir(filepath.Join(workingDir, "frontend"), os.ModePerm)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "frontend", "server.js"), nil, 0644)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "frontend", "app.js"), nil, 0644)).To(Succeed())
			Expect(os.Setenv("BP_NODE_PROJECT_PATH", "frontend")).To(Succeed())
		})

		it.After(func() {
			os.Unsetenv("BP_NODE_PROJECT_PATH")
		})

		it("returns the highest priority file", func() {
			file, err := nodeApplicationFinder.Find(workingDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(file).To(Equal(filepath.Join("frontend", "server.js")))
		})

	}, spec.Sequential())

	context("BP_LAUNCHPOINT is NOT set && when BP_NODE_PROJECT_PATH is NOT set ", func() {
		it.Before(func() {
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "server.js"), nil, 0644)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(workingDir, "app.js"), nil, 0644)).To(Succeed())
		})

		it("returns the highest priority file", func() {
			file, err := nodeApplicationFinder.Find(workingDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(file).To(Equal(filepath.Join("server.js")))
		})

	}, spec.Sequential())

	context("failure cases", func() {
		context("when os.Stat() cannot be performed on the working dir", func() {
			it.Before(func() {
				os.Setenv("BP_LAUNCHPOINT", "something.js")
				Expect(os.Chmod(workingDir, 0000)).To(Succeed())
			})

			it.After(func() {
				os.Unsetenv("BP_LAUNCHPOINT")
				Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
			})

			it("fails with helpful error", func() {
				_, err := nodeApplicationFinder.Find(workingDir)
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
				_, err := nodeApplicationFinder.Find(workingDir)
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		}, spec.Sequential())
	})
}
