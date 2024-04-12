package cmd

import "github.com/spf13/cobra"

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "template-kubernetes-apps-markdown",
		RunE: run,
	}
	cmd.Flags().StringSliceVar(&dirs, "dirs", dirs, "Comma-separated list of directories to template")
	cmd.Flags().StringVar(&pathMatcher, "paths-re", pathMatcher, "Regexp to override certain values. Valid capture groups: cluster, namespace, name")
	cmd.Flags().StringVar(&file, "output", file, "Output filename")
	cmd.Flags().StringVar(&startTag, "start-tag", startTag, "Markdown tag that begins replacement")
	cmd.Flags().StringVar(&endTag, "end-tag", endTag, "Markdown tag that ends replacement")
	cmd.Flags().StringSliceVar(&supportingServices, "supporting-services", supportingServices, "Comma-separated list of supporting service names")
	cmd.Flags().StringSliceVar(&excludedServices, "excluded-services", excludedServices, "Comma-separated list of service names to exclude")
	return cmd
}
