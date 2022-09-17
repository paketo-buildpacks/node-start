package nodestart_test

import (
	"os"
	"path/filepath"
	"testing"

	nodestart "github.com/paketo-buildpacks/node-start"
	"github.com/paketo-buildpacks/packit/v2"
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
		workingDir = t.TempDir()

		Expect(os.WriteFile(filepath.Join(workingDir, "server.js"), nil, 0600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workingDir, "app.js"), nil, 0600)).To(Succeed())
	})

	it("finds the application entrypoint", func() {
		file, err := nodeApplicationFinder.Find(workingDir, "", "")
		Expect(err).NotTo(HaveOccurred())
		Expect(file).To(Equal(filepath.Join("server.js")))
	})

	context("when there is a launchpoint", func() {
		it.Before(func() {
			Expect(os.Mkdir(filepath.Join(workingDir, "src"), os.ModePerm)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(workingDir, "src", "launchpoint.js"), nil, 0600)).To(Succeed())
		})

		it.After(func() {
			Expect(os.RemoveAll(filepath.Join(workingDir, "src"))).To(Succeed())
		})

		context("when the launchpoint file exists", func() {
			it("returns the highest priority file", func() {
				file, err := nodeApplicationFinder.Find(workingDir, "./src/launchpoint.js", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(file).To(Equal(filepath.Join("src", "launchpoint.js")))
			})
		})

		context("when the launchpoint file does not exist", func() {
			it("returns the empty string and no error", func() {
				file, err := nodeApplicationFinder.Find(workingDir, "./no-such-file.js", "")
				Expect(err).To(MatchError(ContainSubstring("expected value derived from BP_LAUNCHPOINT [./no-such-file.js] to be an existing file")))
				Expect(file).To(Equal(""))
			})
		})
	})

	context("when there is a project path", func() {
		it.Before(func() {
			Expect(os.Mkdir(filepath.Join(workingDir, "frontend"), os.ModePerm)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(workingDir, "frontend", "server.js"), nil, 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(workingDir, "frontend", "app.js"), nil, 0600)).To(Succeed())
		})

		it("returns the highest priority file", func() {
			file, err := nodeApplicationFinder.Find(workingDir, "", "frontend")
			Expect(err).NotTo(HaveOccurred())
			Expect(file).To(Equal(filepath.Join("frontend", "server.js")))
		})
	})

	context("when no application can be found", func() {
		it.Before(func() {
			Expect(os.RemoveAll(workingDir)).To(Succeed())
			Expect(os.MkdirAll(workingDir, os.ModePerm)).To(Succeed())
		})

		it("returns a packit failure", func() {
			_, err := nodeApplicationFinder.Find(workingDir, "", "")
			Expect(err).To(MatchError(packit.Fail.WithMessage("could not find app in %s: expected one of server.js | app.js | main.js | index.js", workingDir)))
		})
	})

	context("failure cases", func() {
		context("when the launchpoint cannot be stat'd", func() {
			it.Before(func() {
				Expect(os.Chmod(workingDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
			})

			it("fails with helpful error", func() {
				_, err := nodeApplicationFinder.Find(workingDir, "something.js", "")
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when the working dir cannot be stat'd", func() {
			it.Before(func() {
				Expect(os.Chmod(workingDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
			})

			it("fails with helpful error", func() {
				_, err := nodeApplicationFinder.Find(workingDir, "", "")
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})
	})
}
