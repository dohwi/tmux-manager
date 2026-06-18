# Contributing to tmux-manager

Thanks for your interest! Contributions are welcome — bug reports, feature ideas, and code changes.

## Quick Start

1. **Fork** the repository
2. **Create a branch**: `git checkout -b my-feature`
3. **Make your changes** and add tests if applicable
4. **Run checks**: `go build ./... && go test -race ./...`
5. **Commit** with a clear message (see conventions below)
6. **Push** and open a Pull Request against `main`

## Commit Conventions

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add session sorting by name
fix: resolve pane split direction bug
docs: update keybinding table
refactor: simplify config validation
chore: update dependencies
```

## Pull Requests

- Target the `main` branch
- Keep PRs focused — one concern per PR
- Include tests for new behavior
- Make sure `go vet` and `golangci-lint` pass

## Reporting Issues

- Search existing issues before opening a new one
- Include: tmux version, OS, Go version, and steps to reproduce
- Attach relevant YAML config if applicable (redact any sensitive paths)

## Code Style

- Run `golangci-lint run` before committing
- Follow standard Go formatting (`gofmt`)
- Keep the public API minimal — prefer internal packages

## License

By contributing, you agree that your changes will be licensed under the [MIT License](LICENSE).