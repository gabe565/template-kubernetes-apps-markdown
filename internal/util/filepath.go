package util

import (
	"path/filepath"

	"github.com/gabe565/template-kubernetes-apps-markdown/internal/config"
)

func IsYAMLPath(s string) bool {
	ext := filepath.Ext(s)
	return ext == ".yaml" || ext == ".yml"
}

func MatchPaths(conf *config.Config, path string) (string, string, string) {
	var cluster, namespace, name string
	if conf.PathMatcherRe != nil {
		matches := conf.PathMatcherRe.FindStringSubmatch(path)
		for i, group := range conf.PathMatcherRe.SubexpNames() {
			if i == 0 || len(matches)-1 < i {
				continue
			}

			switch group {
			case "cluster":
				cluster = matches[i]
			case "namespace":
				namespace = matches[i]
			case "name":
				name = matches[i]
			}
		}
	}
	return cluster, namespace, name
}
