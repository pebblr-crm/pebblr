package e2e_test

import (
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var binaryPath string

var _ = BeforeSuite(func() {
	var err error
	binaryPath, err = gexec.Build("github.com/pebblr/pebblr/cmd/pebblr")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func runValidate(configPath string) *gexec.Session {
	cmd := exec.Command(binaryPath, "config", "validate", "--config", configPath)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
	return session
}

func assetPath(parts ...string) string {
	// test/e2e -> test/assets
	base, err := filepath.Abs(filepath.Join("..", "assets"))
	Expect(err).NotTo(HaveOccurred())
	return filepath.Join(append([]string{base}, parts...)...)
}

// ── Valid configs ────────────────────────────────────────────────────────────

var _ = Describe("pebblr config validate", func() {

	Context("with valid configs", func() {
		entries, err := os.ReadDir(filepath.Join("..", "assets", "valid"))
		if err == nil {
			for _, e := range entries {
				name := e.Name()
				It("accepts valid/"+name, func() {
					session := runValidate(assetPath("valid", name))
					Expect(session.ExitCode()).To(Equal(0))
					Expect(string(session.Out.Contents())).To(ContainSubstring("ok"))
				})
			}
		}
	})

	// ── Invalid configs ──────────────────────────────────────────────────────

	Context("with invalid configs", func() {
		entries, err := os.ReadDir(filepath.Join("..", "assets", "invalid"))
		if err == nil {
			for _, e := range entries {
				name := e.Name()
				It("rejects invalid/"+name, func() {
					session := runValidate(assetPath("invalid", name))
					Expect(session.ExitCode()).NotTo(Equal(0))
					Expect(string(session.Err.Contents())).NotTo(BeEmpty())
				})
			}
		}
	})

	// ── Edge cases ───────────────────────────────────────────────────────────

	Context("edge cases", func() {
		It("exits 2 for a missing file", func() {
			cmd := exec.Command(binaryPath, "config", "validate", "--config", "/nonexistent/path.json")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(2))
		})

		It("exits 2 with no arguments", func() {
			cmd := exec.Command(binaryPath)
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(2))
		})

		It("exits 2 for unknown subcommand", func() {
			cmd := exec.Command(binaryPath, "bogus")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(2))
		})

		It("prints help on --help", func() {
			cmd := exec.Command(binaryPath, "--help")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(ContainSubstring("pebblr serve"))
			Expect(string(session.Err.Contents())).To(ContainSubstring("pebblr config"))
		})
	})

	// ── Specific error messages ──────────────────────────────────────────────

	Context("error messages", func() {
		It("reports schema errors for type mismatches", func() {
			session := runValidate(assetPath("invalid", "wrong_type_tenant_name.json"))
			Expect(session.ExitCode()).To(Equal(1))
			stderr := string(session.Err.Contents())
			Expect(stderr).To(ContainSubstring("tenant"))
		})

		It("reports semantic errors for bad transitions", func() {
			session := runValidate(assetPath("invalid", "bad_transition_ref.json"))
			Expect(session.ExitCode()).To(Equal(1))
			stderr := string(session.Err.Contents())
			Expect(stderr).To(ContainSubstring("nonexistent"))
		})

		It("reports missing required section", func() {
			session := runValidate(assetPath("invalid", "missing_rules.json"))
			Expect(session.ExitCode()).To(Equal(1))
			stderr := string(session.Err.Contents())
			Expect(stderr).To(ContainSubstring("rules"))
		})

		It("reports additional properties", func() {
			session := runValidate(assetPath("invalid", "extra_top_level_property.json"))
			Expect(session.ExitCode()).To(Equal(1))
			stderr := string(session.Err.Contents())
			Expect(stderr).To(ContainSubstring("unknown_section"))
		})

		It("reports invalid category enum", func() {
			session := runValidate(assetPath("invalid", "bad_category.json"))
			Expect(session.ExitCode()).To(Equal(1))
			stderr := string(session.Err.Contents())
			Expect(stderr).To(ContainSubstring("category"))
		})

		It("reports unknown title_field", func() {
			session := runValidate(assetPath("invalid", "title_field_unknown.json"))
			Expect(session.ExitCode()).To(Equal(1))
			stderr := string(session.Err.Contents())
			Expect(stderr).To(ContainSubstring("title_field"))
		})

		It("reports submit_required referencing unknown field", func() {
			session := runValidate(assetPath("invalid", "submit_required_unknown_field.json"))
			Expect(session.ExitCode()).To(Equal(1))
			stderr := string(session.Err.Contents())
			Expect(stderr).To(ContainSubstring("ghost_field"))
		})

		It("reports negative max_activities_per_day", func() {
			session := runValidate(assetPath("invalid", "negative_max_activities.json"))
			Expect(session.ExitCode()).To(Equal(1))
			stderr := string(session.Err.Contents())
			Expect(stderr).To(ContainSubstring("max_activities_per_day"))
		})
	})
})
