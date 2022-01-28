package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/onsi/gomega/format"
	"github.com/paketo-buildpacks/occam"
	"github.com/paketo-buildpacks/occam/packagers"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var (
	buildpack           string
	nodeEngineBuildpack string
	watchexecBuildpack  string
	buildpackInfo       struct {
		Buildpack struct {
			ID   string
			Name string
		}
	}
	Config struct {
		NodeEngine string `json:"node-engine"`
		Watchexec  string `json:"watchexec"`
	}
)

func TestIntegration(t *testing.T) {
	format.MaxLength = 0
	Expect := NewWithT(t).Expect

	file, err := os.Open("../integration.json")
	Expect(err).NotTo(HaveOccurred())

	Expect(json.NewDecoder(file).Decode(&Config)).To(Succeed())
	Expect(file.Close()).To(Succeed())

	file, err = os.Open("../buildpack.toml")
	Expect(err).NotTo(HaveOccurred())

	_, err = toml.NewDecoder(file).Decode(&buildpackInfo)
	Expect(err).NotTo(HaveOccurred())
	Expect(file.Close()).To(Succeed())

	root, err := filepath.Abs("./..")
	Expect(err).ToNot(HaveOccurred())

	buildpackStore := occam.NewBuildpackStore()

	libpakBuildpackStore := occam.NewBuildpackStore().WithPackager(packagers.NewLibpak())

	buildpack, err = buildpackStore.Get.
		WithVersion("1.2.3").
		Execute(root)
	Expect(err).ToNot(HaveOccurred())

	nodeEngineBuildpack, err = buildpackStore.Get.
		Execute(Config.NodeEngine)
	Expect(err).ToNot(HaveOccurred())

	watchexecBuildpack, err = libpakBuildpackStore.Get.
		Execute(Config.Watchexec)
	Expect(err).ToNot(HaveOccurred())

	SetDefaultEventuallyTimeout(10 * time.Second)

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Default", testDefault)
	suite("Launchpoint", testLaunchpoint)
	suite("ProjectPath", testProjectPath)
	suite.Run(t)
}
