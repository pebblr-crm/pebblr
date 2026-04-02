package main

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // Ginkgo DSL
	. "github.com/onsi/gomega"    //nolint:revive // Gomega DSL
)

var _ = Describe("readSecret", func() {
	var tmpDir string

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "migrate-test-*")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	It("reads and trims a secret file", func() {
		path := filepath.Join(tmpDir, "dsn")
		Expect(os.WriteFile(path, []byte("  postgres://localhost/test \n"), 0o600)).To(Succeed())

		val, err := readSecret(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(val).To(Equal("postgres://localhost/test"))
	})

	It("returns an error for a missing file", func() {
		_, err := readSecret(filepath.Join(tmpDir, "missing"))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("reading secret file"))
	})

	It("handles an empty file", func() {
		path := filepath.Join(tmpDir, "empty")
		Expect(os.WriteFile(path, []byte(""), 0o600)).To(Succeed())

		val, err := readSecret(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(val).To(BeEmpty())
	})

	It("handles a file with only whitespace", func() {
		path := filepath.Join(tmpDir, "whitespace")
		Expect(os.WriteFile(path, []byte("  \n\t  \n"), 0o600)).To(Succeed())

		val, err := readSecret(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(val).To(BeEmpty())
	})
})

var _ = Describe("run", func() {
	It("fails when DSN file does not exist", func() {
		// The default DSN file path won't exist in a test environment
		err := run()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("reading dsn"))
	})
})
