# Contributing to STACKIT PubSub Go SDK

Thank you for your interest in contributing to the STACKIT PubSub Go SDK! We welcome contributions, bug reports, and suggestions from the community.

Please review this guide before getting started. All contributors are expected to uphold our [Code of Conduct](CODE_OF_CONDUCT.md). If you encounter any issues regarding conduct, please contact `pubsub@stackit.com`.

---

## Reporting Issues & Proposing Features

We use GitHub Issues to track bugs and discuss feature proposals.

### Reporting Bugs

When filing a bug report, please provide as much context as possible to help us diagnose and resolve the issue quickly:

- **Descriptive Title:** Summarize the issue succinctly (e.g., `Subscriber.Pull hangs indefinitely when timeout context is cancelled`).
- **Description:** Clearly state what happened versus what you expected to happen.
- **Minimal Reproducible Example:** Provide a minimal Go snippet demonstrating the problem (client initialization, calls made, error returned).
- **Environment Details:**
  - Go version (e.g., `go version`, or from `mise.toml` / `go.mod`)
  - SDK version (git commit or tag)
  - Operating system and architecture
- **Logs & Error Output:** Include relevant log messages or stack traces. **Important:** Always scrub sensitive information (service account keys, authorization tokens, project secrets) before posting.

### Proposing Features

Before writing code for major changes or new features, please open an issue to discuss the design with the maintainers:

- **Problem & Use Case:** Describe the problem you are solving and why the SDK should support this feature.
- **Proposed Solution:** Outline the suggested API design or method signatures (e.g., how the caller would use it).
- **Alternatives & Compatibility:** Mention any alternative approaches considered, and note if the proposal involves breaking changes.

---

## Pull Request Guidelines

### Workflow & Commits

1. **Branch Naming:** Create a focused branch off `main` with a descriptive prefix:
   - `feat/add-batch-publish-support`
   - `fix/subscriber-ack-race`
   - `docs/clarify-error-handling`
2. **Commit Style:** Use concise commit messages, ideally following [Conventional Commits](https://www.conventionalcommits.org/) (e.g., `feat: ...`, `fix: ...`, `docs: ...`, `chore: ...`).
3. **Scope:** Keep pull requests small and focused on a single concern or bug fix. Avoid bundling unrelated refactorings or dependency changes.

### PR Description Checklist

When opening a PR, include:

- **Summary / Motivation:** Explain why the change is needed and what it accomplishes.
- **Changes Made:** A concise bulleted list of modifications.
- **Linked Issues:** Reference relevant issues (e.g., `Fixes #123` or `Closes #45`).
- **Testing Performed:** Note the tests added or executed to verify the change.
- **Breaking Changes:** Explicitly state if this change breaks backward compatibility and why.

### Review Process

Once your PR is created, reviewers from [CODEOWNERS](CODEOWNERS) are automatically requested. Please address review feedback directly in your branch. Once approved and all CI checks pass, a maintainer will merge the PR.

---

## Testing Requirements

We maintain a high bar for reliability and test coverage:

- **Tests are Mandatory:** Every new feature and bug fix must be accompanied by tests. PRs without tests will not be merged unless the change is purely documentation or tooling.
- **Framework:** The project uses [Ginkgo](https://onsi.github.io/ginkgo/) (v2) and [Gomega](https://onsi.github.io/gomega/) for testing. Tests are located in `pkg/pubsub/*_test.go`.
- **Coverage & Edge Cases:**
  - Verify happy paths as well as error conditions (e.g., `pubsub.APIError` handling, `pubsub.NetworkError`).
  - Tests run with Go's race detector enabled (`-race`); ensure your changes do not introduce data races.
- **Running Tests Locally:**
  - Run the test suite:
    ```bash
    make test
    ```
  - Integration tests interact with STACKIT PubSub APIs and expect configuration variables (e.g., `TOPIC_ID`, `SUBSCRIPTION_ID`, `SERVICE_ACCOUNT_TOKEN`). You can supply these in a local `.env` file at the repository root.

---

## Development Workflow & CI Verification

### Local Commands

The project uses a `Makefile` for standard development tasks:

| Command         | Description                                                                                    |
|-----------------|------------------------------------------------------------------------------------------------|
| `make generate` | Regenerates API client code from `api/openapi.yaml` using `oapi-codegen`.                      |
| `make prepare`  | Tidies Go modules (`go mod tidy`) and automatically fixes linting errors with `golangci-lint`. |
| `make test`     | Runs `make prepare`, followed by the test suite with race detection and coverage reporting.    |

> **Tip:** You can manage tool dependencies (Go `1.25.11`, `golangci-lint`, `oapi-codegen`, `ginkgo`) via [mise](https://mise.jdx.dev/) using the project's `mise.toml` configuration.

### What Runs in CI

Every push and pull request triggers our GitHub Actions CI pipeline (`.github/workflows/ci.yml`), which executes two jobs:

1. **`check-generated-and-lint-code`**:
   - Runs `make generate` and checks `git status`. If any generated files have uncommitted drift, the build fails.
   - Runs `make prepare` to enforce module cleanliness (`go mod tidy`) and linting (`golangci-lint`).
2. **`test-and-resources`**:
   - Sets up the STACKIT CLI and logs in using QA service credentials.
   - Provisions temporary PubSub topics and subscriptions (`./scripts/create-pubsub-resources.sh`).
   - Runs `make test ENABLE_TEST_COVERAGE=true` against the live QA environment.
   - Automatically cleans up and deletes all provisioned resources upon completion (`./scripts/delete-pubsub-resources.sh`).

Before submitting your PR, ensure that `make generate` and `make prepare` leave your working tree clean and that all local tests pass.
