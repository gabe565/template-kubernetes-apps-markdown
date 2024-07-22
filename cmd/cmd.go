package cmd

import (
	"context"

	"github.com/gabe565/template-kubernetes-apps-markdown/internal/config"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "template-kubernetes-apps-markdown",
		RunE: run,
	}
	conf := config.New()
	conf.RegisterFlags(cmd.Flags())
	cmd.SetContext(config.NewContext(context.Background(), conf))
	return cmd
}
