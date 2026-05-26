# Quickstart: M11 Tool Catalog Visibility

## Local Validation

```bash
curl -s http://127.0.0.1:8080/v1/tools/catalog
```

Expected tool names:

- `runtime.get_current_time`
- `workspace.glob`
- `workspace.grep`
- `workspace.read_file`
- `workspace.write_file`
- `workspace.edit`
- `workspace.exec_command`

## Validation Commands

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
(cd docs-site && bun run build)
git diff --check
```
