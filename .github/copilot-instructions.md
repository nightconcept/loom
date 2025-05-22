# AI Project Guidelines (Succinct)

## 1. Preparation
- Review docs/PRD.md (architecture, goals, tech stack, style), docs/digest.txt (current state), docs/TASKS.md (assignments).
- If your task is missing in docs/TASKS.md, add a brief description and date.
- Always review relevant code before changes.

## 2. Implementation Planning
- Before coding, provide: problem summary, high-level solution, steps, and risks.

## 3. Development Workflow
- Present your plan before coding.
- Only address your assigned task; no unrelated refactoring.
- Make minimal, clean, idiomatic changes. Avoid duplication; use helpers/modules.
- Explain significant suggestions. Propose only low-risk refactoring.
- Use only approved dependencies (see docs/PRD.md). No new/updated deps without approval.
- Follow Conventional Commits for user tasks.
- Give clear manual test instructions for changes.

## 4. Folder Structure
- Follow docs/PRD.md structure strictly. No file/folder changes without approval.
- All source code in src/.

## 5. Coding Standards
- **Clarity:** Prioritize readable, well-named, and well-structured code. Comments explain "why" only.
- **Simplicity:** Use the simplest solution. Prefer small, single-purpose functions/files.
- **Concision:** Eliminate repetition and noise. Use idiomatic constructs.
- **Maintainability:** Write code that's easy to modify. Handle errors robustly. Keep tests comprehensive.
- **Consistency:** Match project/language style. Project rules override general best practices.
- Use project formatter. Indent consistently. No spaces before function parentheses.
- Prefer descriptive names; avoid repetition. Use is_/has_ for booleans. Use _ for ignored vars. Avoid single-letter names except for short-lived iterators.
- Each file: header comment with title/purpose. Public APIs: docstrings. Comments explain rationale.
- Type hint where practical (dynamic languages).

## 6. Best Practices (Go)
- **Naming:** Avoid repetition. Use context-appropriate, descriptive names. For test doubles, append 'test' to package name.
- **Shadowing:** Avoid unintentional variable shadowing. Use new names if clarity improves.
- **Util Packages:** Name packages for what they provide, not generic terms like 'util'.
- **Imports:** Group: stdlib, project, (optional) protobuf, (optional) side-effect. Use descriptive import aliases.
- **Error Handling:** Use structured errors. Wrap with %w for unwrapping. Avoid log spam and PII. Propagate init errors to main. Panic only for unrecoverable internal errors.
- **Docs:** Document non-obvious params, concurrency, cleanup, and error conventions. Use godoc formatting.
- **Testing:** Use table-driven tests. Keep helpers focused. Use t.Error for recoverable, t.Fatal for unrecoverable errors. Don't call t.Fatal from goroutines. Use field names in struct literals.
- **Globals:** Avoid mutable global state. If necessary, provide instance-based APIs.

## 7. AI Interaction Protocols
- Your role: Senior Software Engineer. Audience: Mid-level engineers.
- Ask clarifying questions if unclear. Verify facts; don't invent.
- Don't delete/overwrite code unless tasked.
- Report blockers/errors with context and solutions.
- For complex tasks, suggest advanced model if needed.
- Be clear when requirements are met. Mark task complete in docs/TASKS.md (user will update file).



