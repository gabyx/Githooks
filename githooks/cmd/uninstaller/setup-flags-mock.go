//go:build mock

package uninstaller

import (
	cm "github.com/gabyx/githooks/githooks/common"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func setupMockFlags(cmd *cobra.Command, vi *viper.Viper) {
	cmd.PersistentFlags().Bool(
		"stdin", false,
		"Use standard input to read prompt answers.")

	cm.AssertNoErrorPanic(
		vi.BindPFlag("useStdin", cmd.PersistentFlags().Lookup("stdin")))
}
