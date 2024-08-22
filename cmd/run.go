package cmd

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gabe565/template-kubernetes-apps-markdown/internal/config"
	"github.com/gabe565/template-kubernetes-apps-markdown/internal/util"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
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

func run(cmd *cobra.Command, _ []string) error {
	conf, ok := config.FromContext(cmd.Context())
	if !ok {
		panic("command missing config")
	}

	if err := conf.Load(); err != nil {
		return err
	}

	var clusters map[string]Cluster

	var group errgroup.Group
	matchCh := make(chan Match)

	group.Go(func() error {
		defer close(matchCh)
		for _, dir := range conf.Dirs {
			if err := filepath.WalkDir(dir, walkKustomizations(conf, matchCh)); err != nil {
				return err
			}
		}
		for _, dir := range conf.Dirs {
			if err := filepath.WalkDir(dir, walkDirs(conf, matchCh)); err != nil {
				return err
			}
		}
		return nil
	})

	group.Go(func() error {
		clusters = prepareMatches(conf, matchCh)
		return nil
	})

	if err := group.Wait(); err != nil {
		return err
	}

	return templateOutput(conf, clusters)
}

func walkKustomizations(conf *config.Config, matchCh chan Match) fs.WalkDirFunc {
	return func(path string, _ fs.DirEntry, err error) error {
		if err != nil || filepath.Base(path) != "kustomization.yaml" {
			return err
		}

		if conf.ExcludeHidden && strings.Contains(path, string(filepath.Separator)+".") {
			return nil
		}

		docs, err := util.DecodeAll(path)
		if err != nil {
			return fmt.Errorf("unmarshal failed for %q: %w", path, err)
		}

		for _, data := range docs {
			if data, ok := data.(map[string]any); ok {
				apiVersion, _ := data["apiVersion"].(string)
				kind, _ := data["kind"].(string)
				if !strings.HasPrefix(apiVersion, "kustomize.config.k8s.io") || kind != "Kustomization" {
					return nil
				}

				cluster, namespace, name := util.MatchPaths(conf, path)
				if namespace == "" {
					namespace, _ = data["namespace"].(string)
				}

				if resources, ok := data["resources"].([]any); ok {
					for _, resource := range resources {
						resourcePath, ok := resource.(string)
						if !ok {
							continue
						}
						path := filepath.Join(filepath.Dir(path), resourcePath)

						stat, err := os.Stat(path)
						if err != nil {
							if errors.Is(err, os.ErrNotExist) {
								continue
							}
							return err
						}
						if stat.IsDir() {
							continue
						}

						matches, err := getMatches(conf, Match{
							Path:      path,
							Name:      name,
							Cluster:   cluster,
							Namespace: namespace,
						})
						if err != nil {
							return fmt.Errorf("failed to get matches for %q: %w", path, err)
						}

						for _, match := range matches {
							matchCh <- match
						}

						conf.IgnorePaths = append(conf.IgnorePaths, path)
					}
				}
			}
		}
		return nil
	}
}

func walkDirs(conf *config.Config, matchCh chan Match) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !util.IsYAMLPath(path) {
			return err
		}

		if conf.ExcludeHidden && strings.Contains(path, string(filepath.Separator)+".") {
			return nil
		}

		if slices.Contains(conf.IgnorePaths, path) {
			return nil
		}

		matches, err := getMatches(conf, Match{Path: path})
		if err != nil {
			return err
		}

		for _, match := range matches {
			matchCh <- match
		}
		return nil
	}
}

func getMatches(conf *config.Config, match Match) ([]Match, error) { //nolint:gocyclo
	docs, err := util.DecodeAll(match.Path)
	if err != nil {
		return nil, fmt.Errorf("unmarshal failed for %q: %w", match.Path, err)
	}

	matches := make([]Match, 0, len(docs))
	for _, data := range docs {
		data, ok := data.(map[string]any)
		if !ok {
			continue
		}

		m := match
		apiVersion, _ := data["apiVersion"].(string)
		m.Kind, _ = data["kind"].(string)
		metadata, _ := data["metadata"].(map[string]any)
		name, _ := metadata["name"].(string)

		const Appsv1 = "apps/v1"

		switch {
		case slices.Contains(conf.ExcludedServices, name):
			continue
		case apiVersion == Appsv1 && m.Kind == "Deployment":
		case apiVersion == Appsv1 && m.Kind == "StatefulSet":
		case apiVersion == Appsv1 && m.Kind == "DaemonSet":
		case apiVersion == "batch/v1" && m.Kind == "CronJob":
		case strings.HasPrefix(apiVersion, "helm.toolkit.fluxcd.io") && m.Kind == "HelmRelease":
		case strings.HasPrefix(apiVersion, "source.toolkit.fluxcd.io") && m.Kind == "GitRepository" && name != "flux-system":
		case apiVersion == "postgresql.cnpg.io/v1" && m.Kind == "Cluster":
		default:
			continue
		}

		matchedCluster, matchedNamespace, matchedName := util.MatchPaths(conf, m.Path)
		if m.Cluster == "" {
			m.Cluster = matchedCluster
		}
		if m.Namespace == "" {
			m.Namespace = matchedNamespace
			if m.Namespace == "" {
				m.Namespace, _ = metadata["namespace"].(string)
			}
		}
		if m.Name == "" {
			m.Name = matchedName
			if m.Name == "" {
				m.Name = name
			}
		}

		if m.Kind == "GitRepository" {
			if spec, ok := data["spec"].(map[string]any); ok {
				if m.Path, ok = spec["url"].(string); ok {
					m.Path = strings.TrimSuffix(m.Path, ".git")
					if strings.HasPrefix(m.Path, "ssh://git@") {
						m.Path = strings.TrimPrefix(m.Path, "ssh://git@")
						m.Path = "https://" + m.Path
					}
				} else {
					continue
				}
			} else {
				continue
			}
		} else {
			if dir := filepath.Dir(conf.File); dir != "" {
				var err error
				m.Path, err = filepath.Rel(dir, m.Path)
				if err != nil {
					return nil, err
				}
			}
		}

		matches = append(matches, m)
	}
	return matches, nil
}

func prepareMatches(conf *config.Config, matches chan Match) map[string]Cluster {
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

		if slices.Contains(conf.SupportingServices, service.Name) {
			namespace.Supporting[service.Name] = service
		} else {
			namespace.Services[service.Name] = service
		}
	}

	return clusters
}

var (
	ErrNoStartTag = errors.New("no start tag found")
	ErrNoEndTag   = errors.New("no end tag found")
)

func templateOutput(conf *config.Config, clusters map[string]Cluster) error {
	tmpl, err := template.New("").Funcs(funcMap()).Parse(appsTemplate)
	if err != nil {
		return err
	}

	src, err := os.ReadFile(conf.File)
	if err != nil {
		return err
	}

	startIdx := bytes.Index(src, []byte(conf.StartTag))
	if startIdx == -1 {
		return fmt.Errorf("%w: %q in %q", ErrNoStartTag, conf.StartTag, conf.File)
	}

	endIdx := bytes.Index(src, []byte(conf.EndTag))
	if endIdx == -1 {
		return fmt.Errorf("%w: %q in %q", ErrNoEndTag, conf.EndTag, conf.File)
	}

	buf := bytes.NewBuffer(make([]byte, 0, endIdx-startIdx))
	buf.Write(src[:startIdx+len(conf.StartTag)+1])
	if err := tmpl.Execute(buf, clusters); err != nil {
		return err
	}
	buf.Write(src[endIdx:])

	if err := os.WriteFile(conf.File, buf.Bytes(), 0o644); err != nil {
		return err
	}

	return nil
}
