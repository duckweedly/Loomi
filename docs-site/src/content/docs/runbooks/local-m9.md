---
title: M9 Workspace Write Tools 本地验证
description: 本地验证 approval-gated workspace write and edit tools。
---

## 自动化验证

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
cd docs-site && bun run build
git diff --check
```

## 手动 Smoke

1. 触发 `workspace.write_file` 的受控 tool request。
2. 确认 ToolCallCard 显示 approval-required。
3. Deny 一条请求，确认文件未写入。
4. Approve 一条请求，确认只写入一次并显示 terminal result。
5. 触发 `workspace.edit`，确认 `old_text` 精确出现一次时才替换。
6. 刷新页面或重连 SSE，确认历史 replay 顺序不变。
7. 打开浏览器 console，确认没有 error。

## 安全用例

这些输入必须失败，且不能发生 mutation：

```text
../outside.txt
/Users/example/project/file.txt
.env
.ssh/id_ed25519
secrets/token.txt
credentials/prod.json
key.pem
missing-parent/file.txt
```

`workspace.edit` 的这些情况也必须失败且保持原文件不变：

- `old_text` 不存在
- `old_text` 出现多次
- 目标是目录
- 目标是 binary 或无效 UTF-8

M9 不支持 shell、网络、MCP、浏览器自动化、目录创建、二进制写入、长期授权或外部上传。
