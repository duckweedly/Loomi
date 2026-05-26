# Data Model: M8 Safe Workspace Read Tools

## Workspace Read Tool Definition

- `name`: `workspace.glob`, `workspace.grep`, or `workspace.read_file`
- `approval_policy`: `always_required`
- `safety_class`: `workspace_read_only`
- `argument_schema`: tool-specific JSON object
- `result_policy`: bounded, redacted, relative paths only

## Workspace Root

- Local absolute directory configured by the API process.
- M8 development default is the repository working directory.
- All input paths are cleaned, resolved, and checked against this root.

## Tool Arguments

### `workspace.glob`

- `pattern`: required glob pattern, relative to workspace root
- `limit`: optional bounded integer

### `workspace.grep`

- `query`: required plain text or safe regexp according to implementation choice
- `path`: optional relative directory or file path
- `limit`: optional bounded integer

### `workspace.read_file`

- `path`: required relative file path
- `max_bytes`: optional bounded integer

## Tool Results

### Glob Result

- `matches`: bounded relative path list
- `match_count`: returned count
- `truncated`: whether more results were available

### Grep Result

- `matches`: bounded list of `{ path, line, preview }`
- `match_count`: returned count
- `truncated`: whether more matches were available

### Read File Result

- `path`: relative path
- `size_bytes`: file size
- `preview`: bounded UTF-8 text preview
- `truncated`: whether the file was longer than the preview

## Sensitive Path Policy

Inputs are rejected when any normalized path segment or basename matches secret-bearing patterns such as `.env*`, `.ssh`, `.aws`, `secrets`, `credentials`, `*.pem`, `id_rsa*`, and `id_ed25519*`.
