# Contract: Settings and Chat UI

## Settings > Providers

- Detected but disabled Local Codex shows an explicit "Enable for this session" action.
- Enabled supported Local Codex appears in configured providers as local, session-local, redacted, available, and supported.
- Disable removes Local Codex from configured provider candidates.
- No token, key, Authorization header, or auth path is mapped into frontend state.

## Chat Composer

- `execution_state=supported` and `status=available`: no provider unavailable warning, Composer can send.
- `execution_state=unsupported`: show `Local Codex 已启用，但暂不支持执行` and disable send.
- `status=unavailable`: show `Local Codex 登录态不可用，请重新检测或配置 OpenAI-compatible provider` and disable send.

## Timeline and Run Rail

The existing model gateway event projection renders `model_request_started`, `model_output_delta`, `model_output_completed`, `run_completed`, and provider failure events for Local Codex runs without special chat transport.
