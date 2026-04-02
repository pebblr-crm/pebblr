package main

import (
	. "github.com/onsi/ginkgo/v2" //nolint:revive // Ginkgo DSL
	. "github.com/onsi/gomega"    //nolint:revive // Gomega DSL
)

var _ = Describe("run", func() {
	It("returns 2 with no arguments", func() {
		Expect(run([]string{})).To(Equal(2))
	})

	It("returns 0 for --help", func() {
		Expect(run([]string{"--help"})).To(Equal(0))
	})

	It("returns 0 for -h", func() {
		Expect(run([]string{"-h"})).To(Equal(0))
	})

	It("returns 0 for help", func() {
		Expect(run([]string{"help"})).To(Equal(0))
	})

	It("returns 2 for an unknown command", func() {
		Expect(run([]string{"bogus"})).To(Equal(2))
	})

	It("dispatches to serve (which fails without secrets)", func() {
		// serve will fail because there is no config file at the default path,
		// but it should return 1 (runtime error), not 2 (usage error).
		code := run([]string{"serve", "--config", "/nonexistent/config.json"})
		Expect(code).To(Equal(1))
	})

	It("dispatches to config with no subcommand", func() {
		code := run([]string{"config"})
		Expect(code).To(Equal(2))
	})
})
