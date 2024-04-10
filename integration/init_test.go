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
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var settings struct {
	Buildpacks struct {
		NodeStart struct {
			Online string
		}
		NodeEngine struct {
			Online string
		}
		NPMInstall struct {
			Online string
		}
		Watchexec struct {
			Online string
		}
	}

	Extensions struct {
		UbiNodejsExtension struct {
			Online string
		}
	}

	Buildpack struct {
		ID   string
		Name string
	}

	Config struct {
		NodeEngine         string `json:"node-engine"`
		NPMInstall         string `json:"npm-install"`
		Watchexec          string `json:"watchexec"`
		UbiNodejsExtension string `json:"ubi-nodejs-extension"`
	}
}

func TestIntegration(t *testing.T) {
	var docker = occam.NewDocker()

	format.MaxLength = 0
	Expect := NewWithT(t).Expect
	SetDefaultEventuallyTimeout(10 * time.Second)

	file, err := os.Open("../integration.json")
	Expect(err).NotTo(HaveOccurred())

	Expect(json.NewDecoder(file).Decode(&settings.Config)).To(Succeed())
	Expect(file.Close()).To(Succeed())

	file, err = os.Open("../buildpack.toml")
	Expect(err).NotTo(HaveOccurred())

	_, err = toml.NewDecoder(file).Decode(&settings)
	Expect(err).NotTo(HaveOccurred())
	Expect(file.Close()).To(Succeed())

	root, err := filepath.Abs("./..")
	Expect(err).ToNot(HaveOccurred())

	buildpackStore := occam.NewBuildpackStore()
	pack := occam.NewPack()

	builder, err := pack.Builder.Inspect.Execute()
	Expect(err).NotTo(HaveOccurred())

	if builder.BuilderName == "index.docker.io/paketocommunity/builder-ubi-buildpackless-base:latest" {
		settings.Extensions.UbiNodejsExtension.Online, err = buildpackStore.Get.
			Execute(settings.Config.UbiNodejsExtension)
		Expect(err).ToNot(HaveOccurred())
	}

	settings.Buildpacks.NodeStart.Online, err = buildpackStore.Get.
		WithVersion("1.2.3").
		Execute(root)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.NodeEngine.Online, err = buildpackStore.Get.
		Execute(settings.Config.NodeEngine)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.NPMInstall.Online, err = buildpackStore.Get.
		Execute(settings.Config.NPMInstall)
	Expect(err).ToNot(HaveOccurred())
	settings.Buildpacks.Watchexec.Online = settings.Config.Watchexec
	err = docker.Pull.Execute(settings.Buildpacks.Watchexec.Online)
	if err != nil {
		t.Fatalf("Failed to pull %s: %s", settings.Buildpacks.Watchexec.Online, err)
	}

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Default", testDefault)
	suite("Launchpoint", testLaunchpoint)
	suite("ProjectPath", testProjectPath)
	suite("WithNodeModules", testNodeModules)
	suite.Run(t)
}
