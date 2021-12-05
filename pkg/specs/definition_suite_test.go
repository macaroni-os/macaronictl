/*
	Copyright © 2021 RockHopper OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/
package specs_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSolver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Specs defintion Suite")
}
