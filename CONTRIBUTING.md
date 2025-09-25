Claro! Aqui está um CONTRIBUTING.md atualizado e alinhado com o novo README.md do GoBE, destacando práticas modernas, modularidade, MCP tools, testes, arquitetura, segurança e o espírito da comunidade.

---

# Contributing to GoBE

Thank you for your interest in making GoBE better! 🚀  
We welcome contributions from everyone — whether it's fixing bugs, proposing features, improving documentation, or extending the MCP toolset.

## Getting Started

1. **Fork and Clone**
    ```sh
    git clone https://github.com/<your-username>/gobe.git
    cd gobe
    ```

2. **Create a Feature Branch**
    ```sh
    git checkout -b feat/my-feature
    ```

3. **Build & Test Locally**
    ```sh
    make build
    make test
    ```

## How to Contribute

### Code Standards & Architecture

- **Follow the modular architecture**: New features/modules should be encapsulated via interfaces, preferably under `internal/` or `factory/`.
- **MCP Tools**: To add a new MCP tool, use the Registry interface and follow patterns in existing tools (e.g., `system.status`).
- **Thread Safety**: For concurrent code/MCP tools, ensure proper locking (see usage of RWMutex).
- **Go Formatting**: Use `gofmt` and `golangci-lint` on all code.

### Commits & Branches

- Use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) for clear commit messages (e.g., `feat: add new MCP tool`, `fix: auth bug`, `docs: update usage example`).
- Name branches descriptively: `feat/mcp-tool-xyz`, `fix/cli-bug`, etc.
- Open Pull Requests (PRs) against `main` and link related issues if possible.

### Tests

- **Unit and Integration Tests**: All new features, especially MCP tools, must have tests. Use `make test` to run.
- Prefer table-driven tests and check for race conditions in concurrent code.
- PRs without tests (when applicable) may be delayed.

### Documentation

- Update the main [README.md](README.md) or create/update docs under `docs/` for any new feature, endpoint, or tool.
- Document new MCP tools in the relevant section.
- Add usage examples if possible.

### Security

- Never commit secrets, credentials, or sensitive information.
- Use the provided keyring API and recommended security patterns.
- For security-sensitive contributions or reports, email [faelmori@gmail.com](mailto:faelmori@gmail.com) instead of opening a public issue.

### Issues, Discussions & Roadmap

- Open issues to propose major features or discuss architecture before starting large changes.
- Check the [Roadmap](README.md#roadmap) for ideas and ongoing initiatives.
- Suggest new MCP tools, plugins, or monitoring features aligned with project goals.

### Code of Conduct

All contributors must follow our [Code of Conduct](CODE_OF_CONDUCT.md).

We strive for a welcoming, inclusive, and collaborative environment.

---

## Quick Checklist Before Submitting

- [ ] Does your code follow the project’s modular and secure architecture?
- [ ] Did you add/maintain tests?
- [ ] Is the documentation updated?
- [ ] Did you use conventional commits and a clear branch name?
- [ ] Did you run `make build` and `make test` locally?
- [ ] For MCP tools: is it registered via the Registry and thread-safe?
- [ ] No secrets or sensitive data were committed.

If you answered YES to all — you’re ready for a PR!

---

## Need Help?

- Check the [README.md](README.md) and `docs/` for examples.
- Open a [GitHub Discussion](https://github.com/kubex-ecosystem/gobe/discussions) or an Issue.
- Contact the maintainer: [faelmori@gmail.com](mailto:faelmori@gmail.com)

---

Thank you for helping build GoBE and the open source ecosystem! 💚

---

Se quiser adaptar para bilíngue ou incluir exemplos ainda mais práticos (como templates de PR ou exemplos de comandos para MCP tools), só pedir!
