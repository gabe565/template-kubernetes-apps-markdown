package cmd

import "github.com/spf13/cobra"

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "template-kubernetes-apps-markdown",
		RunE: run,
	}
	cmd.Flags().StringArrayVar(&dirs, "dirs", dirs, "Directories to template")
	cmd.Flags().StringVar(&file, "output", file, "Output filename")
	cmd.Flags().StringVar(&startTag, "start-tag", startTag, "Markdown tag that begins replacement")
	cmd.Flags().StringVar(&endTag, "end-tag", endTag, "Markdown tag that ends replacement")
	cmd.Flags().StringArrayVar(&supportingServices, "supporting-services", supportingServices, "Names of supporting services")
	return cmd
}
