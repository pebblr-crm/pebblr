package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // Ginkgo DSL
	. "github.com/onsi/gomega"    //nolint:revive // Gomega DSL
)

// discardLogger returns a logger that writes nowhere, suitable for tests.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

var _ = Describe("runServe", func() {
	It("returns 2 for an invalid flag", func() {
		Expect(runServe([]string{"--bogus"})).To(Equal(2))
	})

	It("returns 1 when config file does not exist", func() {
		code := runServe([]string{"--config", "/nonexistent/config.json"})
		Expect(code).To(Equal(1))
	})

	It("accepts the --auth-provider flag", func() {
		// Will still fail (no config file), but should not return 2 (flag error)
		code := runServe([]string{"--auth-provider", "static", "--config", "/nonexistent/config.json"})
		Expect(code).To(Equal(1))
	})
})

var _ = Describe("readSecretFile", func() {
	var tmpDir string

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "pebblr-test-*")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	It("reads and trims a secret file", func() {
		path := filepath.Join(tmpDir, "secret")
		Expect(os.WriteFile(path, []byte("  my-secret \n"), 0o600)).To(Succeed())

		val, err := readSecretFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(val).To(Equal("my-secret"))
	})

	It("returns an error for a missing file", func() {
		_, err := readSecretFile(filepath.Join(tmpDir, "missing"))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("reading secret file"))
	})
})

var _ = Describe("readOptionalSecret", func() {
	var tmpDir string

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "pebblr-test-*")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	It("reads an existing file", func() {
		path := filepath.Join(tmpDir, "secret")
		Expect(os.WriteFile(path, []byte("value\n"), 0o600)).To(Succeed())

		val, err := readOptionalSecret(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(val).To(Equal("value"))
	})

	It("returns empty string for a missing file", func() {
		val, err := readOptionalSecret(filepath.Join(tmpDir, "missing"))
		Expect(err).NotTo(HaveOccurred())
		Expect(val).To(BeEmpty())
	})
})

var _ = Describe("buildAuthenticator", func() {
	var (
		ctx    context.Context
		logger *slog.Logger
	)

	BeforeEach(func() {
		ctx = context.Background()
		logger = discardLogger()
	})

	It("returns an error for an unknown provider", func() {
		_, _, err := buildAuthenticator(ctx, logger, "unknown", "/tmp", nil)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unknown auth provider"))
	})

	It("blocks static provider in production", func() {
		GinkgoT().Setenv("PEBBLR_ENV", "production")
		_, _, err := buildAuthenticator(ctx, logger, "static", "/tmp", nil)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("not allowed in production"))
	})

	It("blocks demo provider in production", func() {
		GinkgoT().Setenv("PEBBLR_ENV", "production")
		_, _, err := buildAuthenticator(ctx, logger, "demo", "/tmp", nil)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("not allowed in production"))
	})

	It("returns an error when static provider secret file is missing", func() {
		_, _, err := buildAuthenticator(ctx, logger, "static", "/nonexistent", nil)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("reading jwt secret"))
	})

	It("returns an error when azuread tenant ID file is missing", func() {
		_, _, err := buildAuthenticator(ctx, logger, "azuread", "/nonexistent", nil)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("reading Azure AD tenant ID"))
	})

	Context("with valid secret files", func() {
		var tmpDir string

		BeforeEach(func() {
			var err error
			tmpDir, err = os.MkdirTemp("", "pebblr-auth-test-*")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			Expect(os.RemoveAll(tmpDir)).To(Succeed())
		})

		It("creates a static authenticator", func() {
			Expect(os.WriteFile(filepath.Join(tmpDir, "jwt-secret"), []byte("test-secret"), 0o600)).To(Succeed())

			a, demoHandler, err := buildAuthenticator(ctx, logger, "static", tmpDir, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(a).NotTo(BeNil())
			Expect(demoHandler).To(BeNil())
		})

		It("creates a demo authenticator", func() {
			Expect(os.WriteFile(filepath.Join(tmpDir, "demo-signing-key"), []byte("key"), 0o600)).To(Succeed())

			a, demoHandler, err := buildAuthenticator(ctx, logger, "demo", tmpDir, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(a).NotTo(BeNil())
			Expect(demoHandler).NotTo(BeNil())
		})
	})
})
