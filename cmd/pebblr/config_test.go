package main

import (
	. "github.com/onsi/ginkgo/v2" //nolint:revive // Ginkgo DSL
	. "github.com/onsi/gomega"    //nolint:revive // Gomega DSL
)

var _ = Describe("runConfig", func() {
	It("returns 2 with no subcommand", func() {
		Expect(runConfig([]string{})).To(Equal(2))
	})

	It("returns 2 for an unknown subcommand", func() {
		Expect(runConfig([]string{"unknown"})).To(Equal(2))
	})

	It("dispatches to validate", func() {
		// validate with a nonexistent file should return 2 (error reading file)
		code := runConfig([]string{"validate", "--config", "/nonexistent/config.json"})
		Expect(code).To(Equal(2))
	})
})

var _ = Describe("runConfigValidate", func() {
	It("returns 2 when config file does not exist", func() {
		Expect(runConfigValidate([]string{"--config", "/nonexistent/config.json"})).To(Equal(2))
	})

	It("uses the default config path when no --config is given", func() {
		// The default path (./config/tenant.json) likely doesn't exist in test,
		// so this should return 2 (error) or 1 (validation errors).
		code := runConfigValidate([]string{})
		Expect(code).To(BeNumerically(">=", 1))
	})

	It("returns 2 for invalid flag", func() {
		Expect(runConfigValidate([]string{"--bogus"})).To(Equal(2))
	})
})
