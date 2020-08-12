package nodestart_test

import (
	nodestart "github.com/paketo-buildpacks/node-start"
	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir string
		detect     packit.DetectFunc
	)

	it.Before(func() {
		var err error
		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		Expect(ioutil.WriteFile(filepath.Join(workingDir, "server.js"), nil, os.ModePerm)).To(Succeed())

		detect = nodestart.Detect()
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("when there is at least one *.js file in the working directory", func() {
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
		})
	})
	context("when there are no *.js files in the working directory", func() {
		it.Before(func() {
			os.RemoveAll(filepath.Join(workingDir, "server.js"))
		})
		it("fails detection", func() {
			_, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).To(MatchError(packit.Fail))
		})
	})

	context("failure cases", func() {
		context("when file glob fails", func() {
			it("fails with helpful error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: `\`,
				})
				Expect(err).To(MatchError(ContainSubstring("file glob function failed")))
			})
		})
	})
}
