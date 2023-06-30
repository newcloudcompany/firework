package cli

import (
	"github.com/jlkiri/firework/cmd/cleanup"
	"github.com/jlkiri/firework/cmd/connect"
	"github.com/jlkiri/firework/cmd/start"
	"github.com/spf13/cobra"
)

func AddCommands(cmd *cobra.Command) {
	cmd.AddCommand(start.NewStartCommand())
	cmd.AddCommand(connect.NewConnectCommand())
	cmd.AddCommand(cleanup.NewCleanupCommand())
}
