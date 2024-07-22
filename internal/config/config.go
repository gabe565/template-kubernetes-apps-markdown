package config

import "regexp"

type Config struct {
	File               string
	Dirs               []string
	pathMatcher        string
	PathMatcherRe      *regexp.Regexp
	StartTag           string
	EndTag             string
	SupportingServices []string
	ExcludedServices   []string
	ExcludeHidden      bool
}

func New() *Config {
	return &Config{
		File:     "README.md",
		Dirs:     []string{"."},
		StartTag: "<!-- Begin apps section -->",
		EndTag:   "<!-- End apps section -->",
		SupportingServices: []string{
			"postgresql",
			"redis",
			"mariadb",
			"mongodb",
		},
		ExcludedServices: []string{
			"helm-controller",
			"kustomize-controller",
			"notification-controller",
			"source-controller",
		},
	}
}

func (c *Config) Load() error {
	if c.pathMatcher != "" {
		var err error
		c.PathMatcherRe, err = regexp.Compile(c.pathMatcher)
		if err != nil {
			return err
		}
	}

	return nil
}
