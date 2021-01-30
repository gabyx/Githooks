// +build !mock

package installer

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func setupMockFlags(rootCmd *cobra.Command, vi *viper.Viper) {}
