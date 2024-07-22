# Template Kubernetes Apps Markdown

A Go command that automatically generates a repo listing for Kubernetes GitOps repos

## Installation

This repository is intended to be used with [Pre-Commit](https://pre-commit.com).

Example `.pre-commit-config.yaml`:
```yaml
  - repo: https://github.com/gabe565/template-kubernetes-apps-markdown
    rev: ''  # Use the sha / tag you want to point at
    hooks:
      - id: template
        # args:
        #   - --dirs=kubernetes
        #   - --paths-re=^kubernetes/(?P<cluster>.+?)/
        #   - --supporting-services=borgmatic,postgresql,redis,mariadb
```

After adding the hook, your `README.md` will need the following lines added where you want the table to be generated:
```markdown
<!-- Begin apps section -->
<!-- End apps section -->
```

When the hook gets run, the data between those tags will be replaced with a table.

## Usage

See [docs](docs/template-kubernetes-apps-markdown.md) for available flags.

## Example

This repository is used in [gabe565/home-ops](https://github.com/gabe565/home-ops). See the [repo index](https://github.com/gabe565/home-ops#repo-index) section.
