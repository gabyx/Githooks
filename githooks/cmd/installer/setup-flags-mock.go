// +build mock

package installer

import (
	cm "github.com/gabyx/githooks/githooks/common"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func setupMockFlags(rootCmd *cobra.Command, vi *viper.Viper) {
	rootCmd.PersistentFlags().Bool(
		"stdin", false,
		"Use standard input to read prompt answers.")

	cm.AssertNoErrorPanic(
		vi.BindPFlag("useStdin", rootCmd.PersistentFlags().Lookup("stdin")))
}
