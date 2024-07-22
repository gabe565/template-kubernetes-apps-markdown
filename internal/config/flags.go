package config

import "github.com/spf13/pflag"

func (c *Config) RegisterFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&c.Dirs, "dirs", c.Dirs, "Comma-separated list of directories to template")
	fs.StringVar(&c.pathMatcher, "paths-re", c.pathMatcher, "Regexp to override certain values. Valid capture groups: cluster, namespace, name")
	fs.StringVar(&c.File, "output", c.File, "Output filename")
	fs.StringVar(&c.StartTag, "start-tag", c.StartTag, "Markdown tag that begins replacement")
	fs.StringVar(&c.EndTag, "end-tag", c.EndTag, "Markdown tag that ends replacement")
	fs.StringSliceVar(&c.SupportingServices, "supporting-services", c.SupportingServices, "Comma-separated list of supporting service names")
	fs.StringSliceVar(&c.ExcludedServices, "excluded-services", c.ExcludedServices, "Comma-separated list of service names to exclude")
	fs.BoolVar(&c.ExcludeHidden, "exclude-hidden", c.ExcludeHidden, "Excludes hidden files")
}
