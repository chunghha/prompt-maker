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

