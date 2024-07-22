package cmd

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

var (
	file               = "README.md"
	dirs               = []string{"."}
	pathMatcher        string
	pathMatcherRe      *regexp.Regexp
	startTag           = "<!-- Begin apps section -->"
	endTag             = "<!-- End apps section -->"
	supportingServices = []string{
		"postgresql",
		"redis",
		"mariadb",
		"mongodb",
	}
	excludedServices = []string{
		"helm-controller",
		"kustomize-controller",
		"notification-controller",
		"source-controller",
	}
	excludeHidden bool
)

//go:embed apps.html.tmpl
var appsTemplate string

type Cluster struct {
	Name       string
	Namespaces map[string]Namespace
}

type Namespace struct {
	Name       string
	Services   map[string]Match
	Supporting map[string]Match
}

type Match struct {
	Kind      string
	Path      string
	Name      string
	Cluster   string
	Namespace string
}

func run(cmd *cobra.Command, args []string) error {
	if pathMatcher != "" {
		var err error
		if pathMatcherRe, err = regexp.Compile(pathMatcher); err != nil {
			return err
		}
	}

	var clusters map[string]Cluster

	var group errgroup.Group
	matchCh := make(chan Match)

	group.Go(func() error {
		defer close(matchCh)
		for _, dir := range dirs {
			if err := filepath.Walk(dir, walkFunc(matchCh)); err != nil {
				return err
			}
		}
		return nil
	})

	group.Go(func() error {
		clusters = prepareMatches(matchCh)
		return nil
	})

	if err := group.Wait(); err != nil {
		return err
	}

	return templateOutput(clusters)
}

func walkFunc(matchCh chan Match) filepath.WalkFunc {
	outputSubdirCount := strings.Count(file, string(os.PathSeparator))
	outputPathPrefix := strings.Repeat(".."+string(os.PathSeparator), outputSubdirCount)

	return func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if excludeHidden && strings.Contains(path, string(filepath.Separator)+".") {
			return nil
		}

		ext := filepath.Ext(path)
		if info.IsDir() || (ext != ".yaml" && ext != ".yml") {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		decoder := yaml.NewDecoder(f)
		for {
			var data any
			if err := decoder.Decode(&data); err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return fmt.Errorf("unmarshal failed for %q: %w", path, err)
			}

			if data, ok := data.(map[string]any); ok {
				apiVersion, _ := data["apiVersion"].(string)
				kind, _ := data["kind"].(string)
				metadata, _ := data["metadata"].(map[string]any)
				name, _ := metadata["name"].(string)

				switch {
				case slices.Contains(excludedServices, name):
					continue
				case apiVersion == "apps/v1" && kind == "Deployment":
				case apiVersion == "apps/v1" && kind == "StatefulSet":
				case apiVersion == "batch/v1" && kind == "CronJob":
				case strings.HasPrefix(apiVersion, "helm.toolkit.fluxcd.io") && kind == "HelmRelease":
				case strings.HasPrefix(apiVersion, "source.toolkit.fluxcd.io") && kind == "GitRepository" && name != "flux-system":
				case apiVersion == "postgresql.cnpg.io/v1" && kind == "Cluster":
				default:
					continue
				}

				namespace, _ := metadata["namespace"].(string)
				var cluster string

				if pathMatcherRe != nil {
					matches := pathMatcherRe.FindStringSubmatch(path)
					for i, group := range pathMatcherRe.SubexpNames() {
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

				if kind == "GitRepository" {
					if spec, ok := data["spec"].(map[string]any); ok {
						if path, ok = spec["url"].(string); ok {
							path = strings.TrimSuffix(path, ".git")
							if strings.HasPrefix(path, "ssh://git@") {
								path = strings.TrimPrefix(path, "ssh://git@")
								path = "https://" + path
							}
						} else {
							continue
						}
					} else {
						continue
					}
				} else {
					path = filepath.Join(outputPathPrefix, path)
				}

				matchCh <- Match{
					Kind:      kind,
					Path:      path,
					Name:      name,
					Cluster:   cluster,
					Namespace: namespace,
				}
			}
		}
	}
}

func prepareMatches(matches chan Match) map[string]Cluster {
	clusters := make(map[string]Cluster)

	for service := range matches {
		cluster, ok := clusters[service.Cluster]
		if !ok {
			cluster = Cluster{
				Name:       service.Cluster,
				Namespaces: make(map[string]Namespace),
			}
			clusters[cluster.Name] = cluster
		}

		namespace, ok := cluster.Namespaces[service.Namespace]
		if !ok {
			namespace = Namespace{
				Name:       service.Namespace,
				Services:   make(map[string]Match),
				Supporting: make(map[string]Match),
			}
			cluster.Namespaces[namespace.Name] = namespace
		}

		if slices.Contains(supportingServices, service.Name) {
			namespace.Supporting[service.Name] = service
		} else {
			namespace.Services[service.Name] = service
		}
	}

	return clusters
}

func templateOutput(clusters map[string]Cluster) error {
	tmpl, err := template.New("").Funcs(funcMap()).Parse(appsTemplate)
	if err != nil {
		return err
	}

	src, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	startIdx := bytes.Index(src, []byte(startTag))
	if startIdx == -1 {
		return fmt.Errorf("no start tag %q in %q", startTag, file)
	}

	endIdx := bytes.Index(src, []byte(endTag))
	if endIdx == -1 {
		return fmt.Errorf("no end tag %q in %q", endTag, file)
	}

	buf := bytes.NewBuffer(make([]byte, 0, endIdx-startIdx))
	buf.Write(src[:startIdx+len(startTag)+1])
	if err := tmpl.Execute(buf, clusters); err != nil {
		return err
	}
	buf.Write(src[endIdx:])

	if err := os.WriteFile(file, buf.Bytes(), 0o644); err != nil {
		return err
	}

	return nil
}
