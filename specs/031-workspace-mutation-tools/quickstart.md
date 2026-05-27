# Quickstart: M23 Workspace Mutation Tools

## Focused Backend Tests

```bash
go test ./internal/runtime -run 'TestWorkspaceMutation|TestToolBrokerExecutesWorkspaceMutation'
go test ./internal/productdata -run 'TestWorkspaceMutation|TestWorkspaceTools'
go test ./internal/httpapi -run TestM23WorkspaceMutationToolsSmoke
```

Expected evidence:

1. `workspace.write_file` creates exactly one new file after approval.
2. The same request does not write before approval.
3. Existing target, traversal, absolute escape, symlink escape, sensitive path, and too-large content fail without mutation.
4. `workspace.edit` replaces exactly one existing text occurrence after approval.
5. Ambiguous/missing old text fails without mutation.
6. Terminal/stopped/denied paths do not mutate files.

## Frontend Checks

```bash
bun test --cwd web ./src/components/SettingsView.tools.test.tsx ./src/components/RunRail.runtime.test.ts
bun run --cwd web build
```

Expected evidence:

1. Settings Tools shows `workspace.write_file` and `workspace.edit` as workspace scoped, approval required, write capable, and high risk.
2. RunRail shows mutation lifecycle rows without raw file contents or host absolute root paths.

## Full Validation Target

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Browser Smoke

Run the web shell and verify the seeded/mock Settings Tools and RunRail mutation rows:

```bash
cd web
bun run dev --host 127.0.0.1 --port 5173
```

Open `http://127.0.0.1:5173/`, inspect Settings > Tools or Run details, and confirm no console errors.
