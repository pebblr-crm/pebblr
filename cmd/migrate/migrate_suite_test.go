package main

import (
	"testing"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // Ginkgo DSL
	. "github.com/onsi/gomega"    //nolint:revive // Gomega DSL
)

func TestMigrate(t *testing.T) { //nolint:paralleltest // Ginkgo manages parallelism
	RegisterFailHandler(Fail)
	RunSpecs(t, "Migrate CLI Suite")
}
