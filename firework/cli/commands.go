package cli

import (
	"github.com/jlkiri/firework/cmd/connect"
	"github.com/jlkiri/firework/cmd/start"
	"github.com/jlkiri/firework/cmd/status"
	"github.com/jlkiri/firework/cmd/stop"
	"github.com/spf13/cobra"
)

func AddCommands(cmd *cobra.Command) {
	cmd.AddCommand(start.NewStartCommand())
	cmd.AddCommand(connect.NewConnectCommand())
	cmd.AddCommand(stop.NewStopCommand())
	cmd.AddCommand(status.NewStatusCommand())
}
