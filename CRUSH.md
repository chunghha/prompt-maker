# Go Project Commands and Style Guide

## Commands

- **Build:** `task build`
- **Test:** `task test`
- **Test a specific test:** `go test -v ./path/to/your_test.go -run TestYourFunction`
- **Lint:** `task lint`
- **Format:** `task format`
- **Run web server:** `task run:web`

## Code Style

- **Formatting:** Use `gofmt` and `goimports`. Run `task format` before committing.
- **Linting:** Adhere to `.golangci.yml`. Run `task lint`. No linting errors should be present in commits.
- **Imports:** Group imports in the following order: standard, third-party, and local.
- **Naming Conventions:**
    - Use camelCase for variables and functions.
    - Use PascalCase for public functions and structs.
    - Keep variable names short but descriptive.
- **Error Handling:**
    - Handle errors explicitly; do not discard them with `_`.
    - Use `errors.Is()` and `errors.As()` for checking error types.
- **Comments:**
    - Comment on complex logic, not on what the code does.
    - Use `//` for single-line and `/* */` for multi-line comments.
- **Testing:**
    - Write tests for all new features and bug fixes.
    - Use the `testing` package and `testify/assert` for assertions.
    - Keep tests in the same package as the code they are testing, with the `_test.go` suffix.
- **Frontend:**
    - Use `pnpm` for package management.
    - Build CSS with `pnpm run css:build`.

## Commit Messages

- Use lowercase for the commit type (e.g., `feat:`, `fix:`, `refactor:`).
