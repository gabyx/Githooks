// +build !mock

package uninstaller

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func setupMockFlags(cmd *cobra.Command, vi *viper.Viper) {}
