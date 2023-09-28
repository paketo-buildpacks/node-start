package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testDefault(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker

		pullPolicy              = "never"
		extenderBuildStr        = ""
		extenderBuildStrEscaped = ""
	)

	it.Before(func() {
		pack = occam.NewPack().WithVerbose().WithNoColor()
		docker = occam.NewDocker()

		if settings.Extensions.UbiNodejsExtension.Online != "" {
			pullPolicy = "always"
			extenderBuildStr = "[extender (build)] "
			extenderBuildStrEscaped = `\[extender \(build\)\] `
		}
	})

	context("when building a default app", func() {
		var (
			image     occam.Image
			container occam.Container

			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("builds and runs successfully", func() {
			var err error
			source, err = occam.Source(filepath.Join("testdata", "default"))
			Expect(err).NotTo(HaveOccurred())

			var logs fmt.Stringer
			image, logs, err = pack.Build.
				WithPullPolicy(pullPolicy).
				WithExtensions(
					settings.Extensions.UbiNodejsExtension.Online,
				).
				WithBuildpacks(
					settings.Buildpacks.NodeEngine.Online,
					settings.Buildpacks.NodeStart.Online,
				).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			container, err = docker.Container.Run.
				WithEnv(map[string]string{"PORT": "8080"}).
				WithPublish("8080").
				WithPublishAll().
				Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(container).Should(Serve(ContainSubstring("hello world")))

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s%s \d+\.\d+\.\d+`, extenderBuildStrEscaped, settings.Buildpack.Name)),
				extenderBuildStr+"  Assigning launch processes:",
				extenderBuildStr+"    web (default): node server.js",
			))
		})

		context("when BP_LIVE_RELOAD_ENABLED=true", func() {
			var noReloadContainer occam.Container

			it.After(func() {
				Expect(docker.Container.Remove.Execute(noReloadContainer.ID)).To(Succeed())
			})

			it("adds a reloadable process type as the default process", func() {
				var err error
				source, err = occam.Source(filepath.Join("testdata", "default"))
				Expect(err).NotTo(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.Build.
					WithPullPolicy(pullPolicy).
					WithExtensions(
						settings.Extensions.UbiNodejsExtension.Online,
					).
					WithBuildpacks(
						settings.Buildpacks.NodeEngine.Online,
						settings.Buildpacks.Watchexec.Online,
						settings.Buildpacks.NodeStart.Online,
					).
					WithEnv(map[string]string{
						"BP_LIVE_RELOAD_ENABLED": "true",
					}).
					Execute(name, source)
				Expect(err).ToNot(HaveOccurred(), logs.String)

				container, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					WithPublishAll().
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(Serve(ContainSubstring("hello world")).OnPort(8080))

				noReloadContainer, err = docker.Container.Run.
					WithEnv(map[string]string{"PORT": "8080"}).
					WithPublish("8080").
					WithPublishAll().
					WithEntrypoint("no-reload").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(noReloadContainer).Should(Serve(ContainSubstring("hello world")).OnPort(8080))

				Expect(logs).To(ContainLines(
					MatchRegexp(fmt.Sprintf(`%s%s \d+\.\d+\.\d+`, extenderBuildStrEscaped, settings.Buildpack.Name)),
					extenderBuildStr+"  Assigning launch processes:",
					extenderBuildStr+"    web (default): watchexec --restart --watch /workspace --shell none -- node server.js",
					extenderBuildStr+"    no-reload:     node server.js",
				))
			})
		})
	})
}
