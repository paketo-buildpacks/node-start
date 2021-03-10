package nodestart_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitNodeStart(t *testing.T) {
	suite := spec.New("node-start", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Build", testBuild)
	suite("Detect", testDetect)
	suite("ApplicationFinder", testNodeApplicationFinder)
	suite.Run(t)
}
