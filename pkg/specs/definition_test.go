/*
	Copyright © 2021 Macaroni OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/
package specs_test

import (
	"os"
	"strings"

	. "github.com/funtoo/macaronictl/pkg/specs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v "github.com/spf13/viper"
)

var _ = Describe("Specs Test", func() {

	Context("Config1", func() {
		os.Setenv("MACARONICTL_GENERAL__DEBUG", "true")
		config := NewMacaroniCtlConfig(v.New())
		// Set env variable
		config.Viper.SetEnvPrefix(MACARONICTL_ENV_PREFIX)
		config.Viper.BindEnv("config")
		config.Viper.SetDefault("config", "")
		config.Viper.SetDefault("etcd-config", false)

		config.Viper.AutomaticEnv()

		// Create EnvKey Replacer for handle complex structure
		replacer := strings.NewReplacer(".", "__")
		config.Viper.SetEnvKeyReplacer(replacer)

		err := config.Unmarshal()

		It("Convert env1", func() {

			Expect(err).Should(BeNil())
			Expect(config.GetGeneral().Debug).To(Equal(true))
		})

	})

})
