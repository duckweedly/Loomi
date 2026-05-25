# Contract: Settings Provider UI

Settings > Providers keeps the manual Local Provider Autodetect button.

When detection returns Local Codex `available`, the local provider card shows an explicit enable action. Enabled local providers show a disable action, session-local scope, redacted credential reference, and unsupported execution status.

Chat Composer treats enabled-but-unsupported local providers as not ready for sending and keeps the provider warning visible.

The UI must never render token, refresh token, API key, private path, or raw auth content.
