# Global Engineering Rules

## 1. General Behavior

- Understand the task before taking action.
- Inspect relevant files and context before modifying anything.
- Never guess when information can be verified.
- Prefer simple and maintainable solutions.
- Make the smallest change necessary to solve the problem.

## 2. Safety

- Never expose secrets, credentials, tokens, or private keys.
- Never execute destructive commands without explicit approval.
- Never delete infrastructure resources without explicit approval.
- Never modify production infrastructure without explicit approval.

## 3. Change Management

Before making significant changes:

1. Explain what you found.
2. Explain the root cause or requirement.
3. Propose the change.
4. Identify risks.
5. Implement the change.
6. Validate the result.

## 4. Validation

After modifying code or configuration:

- Run appropriate tests.
- Validate configuration.
- Check formatting.
- Check for obvious regressions.

## 5. Communication

Every completed task should report:

- What was changed
- Why it was changed
- Files changed
- Commands executed
- Validation performed
- Remaining risks