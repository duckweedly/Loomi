# Data Model: M9 Safe Workspace Write Tools

## Workspace Write Tool Definition

- `name`: `workspace.write_file` or `workspace.edit`
- `approval_policy`: `always_required`
- `safety_class`: `workspace_write`
- `argument_schema`: tool-specific JSON object
- `result_policy`: bounded, redacted, relative paths only

## Tool Arguments

### `workspace.write_file`

- `path`: required relative file path
- `content`: required UTF-8 text content

### `workspace.edit`

- `path`: required relative existing file path
- `old_text`: required exact text to replace
- `new_text`: required replacement text

## Tool Results

### Write File Result

- `path`: relative path
- `bytes_written`: number of bytes written
- `created`: whether the file did not exist before execution
- `truncated`: always false for accepted bounded writes

### Edit Result

- `path`: relative path
- `replacements`: `1`
- `bytes_before`: original file size
- `bytes_after`: new file size

## Sensitive Path Policy

M9 reuses and hardens M8 sensitive path rejection. Sensitive targets are denied before mutation even when they are inside the workspace root.
