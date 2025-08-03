- [x] Centralize error handling with a custom middleware.
- [x] Optimize `markdownToHTML` by initializing Goldmark once.
- [x] Refactor `handleIndex` to be model-agnostic.
- [x] Improve `handleClear` to return `204 No Content`.

### TUI Enhancements

- [x] Refactor `Update` function into smaller, state-specific functions.
- [x] Centralize all style definitions into the `Styles` struct.
- [x] Improve readability of `handleKeyMsg` by refactoring complex conditionals.
- [x] Abstract `gemini.ModelOption` to a more generic interface.
- [x] Remove `isPromptCrafted` flag and simplify logic by checking `craftedPrompt`.

### Suggested Enhancements

- [x] Refactor the massive `internal/tui/tui.go` file into smaller, more manageable components.
- [x] Add CLI flags for `--model`, `--temperature`, and `--history`.
- [ ] Improve TUI error messages to be more user-friendly.
- [ ] Improve the web UI with a modern CSS framework.
- [ ] Add unit tests for the web handlers.
- [ ] Add unit tests for the TUI.
- [ ] Add a CI pipeline to run tests on every push.
- [ ] Add a release pipeline to automate the creation of GitHub releases.
- [ ] Dynamically load models from the Gemini API instead of using a hardcoded list.
