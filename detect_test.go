package nodestart_test

import (
	"io/ioutil"
	"os"
	"testing"

	nodestart "github.com/paketo-buildpacks/node-start"
	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"

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

		Expect(ioutil.WriteFile(filepath.Join(workingDir, "server.js"), nil, 0644)).To(Succeed())

		detect = nodestart.Detect()
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("when there is at least one *.js file in the working directory", func() {
		it.Focus("detects", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "node",
						Metadata: map[string]interface{}{
							"launch": true, // build or launch?
						},
					},
				},
			}))
		})
	})
	context("when there are no *.js files in the working directory", func() {
		it("fails detection", func() {
			_, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).To(MatchError(packit.Fail))
		})
	})
}
