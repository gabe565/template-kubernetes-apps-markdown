## template-kubernetes-apps-markdown



```
template-kubernetes-apps-markdown [flags]
```

### Options

```
      --dirs strings                  Comma-separated list of directories to template (default [.])
      --end-tag string                Markdown tag that ends replacement (default "<!-- End apps section -->")
      --exclude-hidden                Excludes hidden files
      --excluded-services strings     Comma-separated list of service names to exclude (default [helm-controller,kustomize-controller,notification-controller,source-controller])
  -h, --help                          help for template-kubernetes-apps-markdown
      --output string                 Output filename (default "README.md")
      --paths-re string               Regexp to override certain values. Valid capture groups: cluster, namespace, name
      --start-tag string              Markdown tag that begins replacement (default "<!-- Begin apps section -->")
      --supporting-services strings   Comma-separated list of supporting service names (default [postgresql,redis,mariadb,mongodb])
```

