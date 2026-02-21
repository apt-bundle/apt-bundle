Perform a thorough, senior-engineer-level code review of the entire codebase. Work methodically through every source file. For each issue found, cite the file path and line number(s).

## Review dimensions

Work through every dimension below. Do not skip any.

### 1. Correctness & logic bugs
- Off-by-one errors, nil/zero-value mishandling, incorrect boolean logic.
- Missing or incorrect error propagation (especially in Go: unchecked errors, shadowed `err`).
- Race conditions or unsafe concurrent access to shared state.
- Edge cases: empty inputs, boundary values, unicode, very large inputs.

### 2. Idiomatic Go
- Follow Go conventions: naming (MixedCaps, not underscores), receiver names, package naming.
- Prefer `errors.Is` / `errors.As` over string comparison on errors.
- Use `fmt.Errorf("context: %w", err)` for wrapping; avoid bare `fmt.Errorf("%s", err)`.
- Avoid stuttering in exported names (e.g., `repo.RepoManager` → `repo.Manager`).
- Prefer table-driven tests; avoid deep nesting in test functions.
- Use `t.Helper()` in test helpers.
- Return early to reduce indentation (guard clauses).
- Avoid unnecessary `else` after a return.
- Prefer `strings.Contains` / `strings.HasPrefix` over manual index checks.
- Avoid `init()` functions unless truly necessary.

### 3. Design & architecture
- Single Responsibility: does each package/type do one thing well?
- Are there God objects or functions that do too much?
- Is there duplicated logic that should be consolidated?
- Are abstractions at the right level — not too abstract, not too concrete?
- Is the dependency graph clean or are there circular/tangled imports?
- Are interfaces defined where they're consumed, not where they're implemented?
- Is there dead code (unused functions, types, constants, or unexported identifiers)?

### 4. Error handling
- Are errors handled at every call site or explicitly ignored with a comment?
- Are sentinel errors or custom error types used consistently?
- Are user-facing error messages clear, actionable, and free of internal jargon?
- Is there any swallowed error (e.g., `_ = someFunc()`) that should be surfaced?

### 5. Security
- Command injection: are any user-supplied strings passed unsanitized to `exec.Command` or shell invocations?
- Path traversal: are file paths validated and cleaned (`filepath.Clean`, `filepath.Rel`)?
- Are secrets, tokens, or credentials ever logged or included in error messages?
- Are URLs validated before use (scheme, host)?
- Are permissions on created files/directories appropriately restrictive?
- Is input from external sources (files, env vars, HTTP) validated before use?

### 6. Documentation & comments
- Are comments accurate and in sync with the code they describe?
- Are there stale TODO/FIXME/HACK comments that should be resolved or removed?
- Do exported functions and types have godoc comments?
- Are misleading or redundant comments present (comments that just restate the code)?

### 7. Testing
- Are there obvious gaps in test coverage for important code paths?
- Do tests actually assert meaningful behavior or just exercise code without checking results?
- Are tests isolated (no shared mutable state, no order dependency)?
- Are test file names and function names descriptive?
- Is there test code that tests implementation details rather than behavior?

### 8. Performance (only if clearly relevant)
- Unnecessary allocations in hot paths (e.g., repeated string concatenation in loops).
- Unbounded growth of slices or maps.
- Unnecessary repeated I/O or network calls that could be batched or cached.

### 9. Miscellaneous
- Are there any `//nolint` directives that suppress warnings without justification?
- Is there inconsistent formatting or style across the codebase?
- Are build tags, module paths, or dependency versions stale?
- Are there any leftover debugging artifacts (print statements, hardcoded test values)?

## Output format

Organize findings by dimension. Within each dimension, list issues sorted by severity (critical → minor). For each issue:

```
**[severity]** `file/path.go:line` — description of the issue and suggested fix.
```

Severity levels: **critical** (bugs, security), **warning** (design, correctness risk), **nit** (style, idiom).

At the end, provide a brief summary: total counts by severity and a top-3 list of the most impactful improvements to make first.
