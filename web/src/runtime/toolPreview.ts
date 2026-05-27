import type { RunEvent } from '../domain'
import type { Locale } from '../i18n'

const redacted = '[redacted]'

const sensitiveKeyPattern = /(^|_|\b)(authorization|api[-_]?key|token|secret|cookie|set[-_]?cookie|path|cwd|stdout|stderr|raw[-_]?body|body|html|headers|content|input|output|diff|snippet|preview[-_]?id)($|_|\b)/i
const pathPattern = /(?:\/(?:Users|home|tmp)\/[^\s"'<>]+|[A-Za-z]:\\[^\s"'<>]+|[^\s"'<>]*\.env[^\s"'<>]*)/g
const pathDetectPattern = /(?:\/(?:Users|home|tmp)\/[^\s"'<>]+|[A-Za-z]:\\[^\s"'<>]+|[^\s"'<>]*\.env[^\s"'<>]*)/
const relativePathPattern = /(^|[\s"'(:])(?:[A-Za-z0-9_.-]+\/)+[A-Za-z0-9_.-]+\.[A-Za-z0-9_-]+/g
const relativePathDetectPattern = /(^|[\s"'(:])(?:[A-Za-z0-9_.-]+\/)+[A-Za-z0-9_.-]+\.[A-Za-z0-9_-]+/
const secretPattern = /(?:Bearer\s+)?sk-[A-Za-z0-9_-]+|Authorization\s*[:=]\s*[^\s"'<>]+|cookie\s*[:=]\s*[^\s"'<>]+/gi
const secretDetectPattern = /(?:Bearer\s+)?sk-[A-Za-z0-9_-]+|Authorization\s*[:=]\s*[^\s"'<>]+|cookie\s*[:=]\s*[^\s"'<>]+/i
const sensitivePairPattern = /\b(token|api_key|apikey|secret|password|credential|session)\b\s*[:=]\s*[^\s"'<>]+/gi
const sensitivePairDetectPattern = /\b(token|api_key|apikey|secret|password|credential|session)\b\s*[:=]\s*[^\s"'<>]+/i
const sensitiveHeaderPattern = /\b(Authorization|Cookie|Set-Cookie)\b\s*[:=]?\s*(?:Bearer\s+)?(?:\[redacted\]|[^\s"'<>]+)?/gi

const toolLabels: Record<string, string> = {
  'workspace.read': 'Read project files',
  'workspace.read_file': 'Read workspace file',
  'workspace.glob': 'Find project files',
  'workspace.grep': 'Search project text',
  'workspace.list_directory': 'Read directory',
  'workspace.tree_summary': 'Summarize directory',
  'workspace.write_file': 'Change workspace files',
  'workspace.edit': 'Change workspace files',
  'workspace.patch_preview': 'Preview workspace patch',
  'workspace.patch_apply': 'Apply workspace patch',
  'web.fetch': 'Visit web page',
  'web.search': 'Search web',
  'lsp.symbols': 'Analyze code',
  'sandbox.exec_command': 'Run sandbox command',
  'sandbox.start_process': 'Start sandbox process',
  'sandbox.continue_process': 'Continue sandbox process',
  'sandbox.terminate_process': 'Terminate sandbox process',
  'runtime.get_current_time': 'Get current time',
  'todo.write': 'Update work plan',
}

const zhToolLabels: Record<string, string> = {
  'workspace.read': '读取项目文件',
  'workspace.read_file': '读取工作区文件',
  'workspace.glob': '查找项目文件',
  'workspace.grep': '搜索项目文本',
  'workspace.list_directory': '读取目录',
  'workspace.tree_summary': '目录概览',
  'workspace.write_file': '修改工作区文件',
  'workspace.edit': '修改工作区文件',
  'workspace.patch_preview': '预览工作区补丁',
  'workspace.patch_apply': '应用工作区补丁',
  'web.fetch': '访问网页',
  'web.search': '搜索网页',
  'lsp.symbols': '分析代码',
  'sandbox.exec_command': '运行沙箱命令',
  'sandbox.start_process': '启动沙箱进程',
  'sandbox.continue_process': '继续沙箱进程',
  'sandbox.terminate_process': '终止沙箱进程',
  'runtime.get_current_time': '获取当前时间',
  'todo.write': '更新工作计划',
}

const keyLabels = {
  zh: {
    query: '搜索词',
    provider: '服务',
    limit: '数量',
    timeout_ms: '超时',
    title: '标题',
    items: '结果',
    todo_items: '待办',
    url: '链接',
    snippet: '摘要',
    result_count: '结果',
    status_code: '状态',
    timezone: '时区',
    iso_time: '时间',
  },
  en: {},
} as const

const allowedKeys = new Set([
  'timezone',
  'iso_time',
  'local_time',
  'query',
  'pattern',
  'limit',
  'line',
  'preview',
  'match_count',
  'total_entries_seen',
  'returned_entries',
  'directories_count',
  'files_count',
  'by_extension',
  'by_kind',
  'largest_files',
  'recent_files',
  'matches',
  'status_code',
  'bytes_read',
  'bytes_written',
  'bytes_before',
  'bytes_after',
  'created',
  'replacements',
  'exit_code',
  'timed_out',
  'stdout',
  'stderr',
  'stdout_truncated',
  'stderr_truncated',
  'process_id',
  'argv_summary',
  'cwd_alias',
  'next_cursor',
  'terminal_summary',
  'stdin_open',
  'input_seq',
  'started_at',
  'updated_at',
  'ended_at',
  'truncated',
  'total',
  'completed_count',
  'in_progress_count',
  'pending_count',
  'items',
  'todo_items',
  'title',
  'url',
  'snippet',
  'provider',
  'result_count',
  'status',
  'server',
  'message',
  'side_effect',
  'operation',
  'changed',
  'diff',
  'snippet',
  'preview_id',
])

function statusCopy(event: RunEvent, locale: Locale = 'en') {
  if (locale === 'zh') {
    if (event.status === 'blocked_on_tool_approval' || event.type.endsWith('.approval_required')) return '等待确认'
    if (event.status === 'failed' || event.type.endsWith('.failed')) return '失败'
    if (event.status === 'cancelled' || event.type.endsWith('.cancelled')) return '已取消'
    if (event.status === 'stopped') return '已停止'
    if (event.status === 'completed' || event.type.endsWith('.succeeded')) return '完成'
    if (event.status === 'running' || event.type.endsWith('.executing') || event.type.endsWith('.requested')) return '运行中'
  } else {
    if (event.status === 'blocked_on_tool_approval' || event.type.endsWith('.approval_required')) return 'waiting for approval'
    if (event.status === 'failed' || event.type.endsWith('.failed')) return 'failed'
    if (event.status === 'cancelled' || event.type.endsWith('.cancelled')) return 'cancelled'
    if (event.status === 'stopped') return 'stopped'
    if (event.status === 'completed' || event.type.endsWith('.succeeded')) return 'completed'
    if (event.status === 'running' || event.type.endsWith('.executing') || event.type.endsWith('.requested')) return 'running'
  }
  return event.status.replaceAll('_', ' ')
}

export function humanToolName(name: string | undefined, locale: Locale = 'en') {
  if (!name) return 'Tool call'
  if (locale === 'zh') {
    if (zhToolLabels[name]) return zhToolLabels[name]
    if (name.startsWith('artifact.')) return '处理产物'
    if (name.startsWith('agent.')) return '协调子任务'
    if (name.startsWith('workspace.')) return '检查工作区'
    if (name.startsWith('lsp.')) return '分析代码'
    if (name.startsWith('web.')) return '使用网页'
    if (name.startsWith('browser.')) return '操作浏览器'
    return '使用工具'
  }
  if (toolLabels[name]) return toolLabels[name]
  if (name.startsWith('artifact.')) return 'Handle artifact'
  if (name.startsWith('agent.')) return 'Coordinate subtasks'
  if (name.startsWith('workspace.')) return 'Inspect workspace'
  if (name.startsWith('lsp.')) return 'Analyze code'
  if (name.startsWith('web.')) return 'Use web'
  if (name.startsWith('browser.')) return 'Use browser'
  return 'Use tool'
}

export function redactPreviewText(value: string) {
  return value
    .replace(pathPattern, redacted)
    .replace(relativePathPattern, `$1${redacted}`)
    .replace(secretPattern, redacted)
    .replace(sensitivePairPattern, redacted)
    .replace(sensitiveHeaderPattern, redacted)
}

function isSensitiveKey(key: string) {
  if (key === 'cwd_alias') return false
  return sensitiveKeyPattern.test(key)
}

function valueLooksSensitive(value: string) {
  return pathDetectPattern.test(value) || relativePathDetectPattern.test(value) || secretDetectPattern.test(value) || sensitivePairDetectPattern.test(value)
}

function urlHost(value: unknown) {
  if (typeof value !== 'string') return ''
  try {
    return new URL(value).hostname.replace(/^www\./, '')
  } catch {
    return ''
  }
}

function truncatePreview(value: string, max = 96) {
  const text = redactPreviewText(value).replace(/\s+/g, ' ').trim()
  return text.length > max ? `${text.slice(0, max - 1)}...` : text
}

function formatWebSearchPreview(value: Record<string, unknown>, locale: Locale = 'en') {
  if (!Array.isArray(value.items)) return ''
  const items = value.items
    .filter((item): item is Record<string, unknown> => typeof item === 'object' && item !== null && !Array.isArray(item))
    .filter((item) => typeof item.title === 'string' || typeof item.url === 'string' || typeof item.snippet === 'string')
  if (items.length === 0) return ''
  const shown = items.slice(0, 3).map((item) => {
    const title = truncatePreview(String(item.title ?? item.url ?? ''), 74)
    const host = urlHost(item.url)
    const snippet = typeof item.snippet === 'string' ? truncatePreview(item.snippet, 70) : ''
    return [title, host ? `(${host})` : '', snippet ? `- ${snippet}` : ''].filter(Boolean).join(' ')
  }).filter(Boolean)
  const rest = items.length - shown.length
  return rest > 0 ? `${shown.join(' · ')} · ${locale === 'zh' ? `另 ${rest} 项` : `${rest} more`}` : shown.join(' · ')
}

function formatValue(key: string, value: unknown, locale: Locale = 'en'): string {
  if (isSensitiveKey(key)) return redacted
  if (Array.isArray(value)) {
    const shown = value.slice(0, 3).map((item) => formatArrayItem(key, item, locale)).filter(Boolean)
    const rest = value.length - shown.length
    return rest > 0 ? `${shown.join(', ')}${shown.length > 0 ? ' · ' : ''}${locale === 'zh' ? `另 ${rest} 项` : `${rest} more`}` : shown.join(', ')
  }
  if (typeof value === 'object' && value !== null) return formatSafeToolPreview(value as Record<string, unknown>, locale)
  const text = String(value)
  return valueLooksSensitive(text) ? redacted : redactPreviewText(text)
}

function formatArrayItem(key: string, item: unknown, locale: Locale = 'en'): string {
  if (typeof item !== 'object' || item === null) return formatValue(key, item, locale)
  const record = item as Record<string, unknown>
  if ('line' in record || 'preview' in record) {
    return [
      record.line !== undefined ? `line ${formatValue('line', record.line, locale)}` : null,
      record.preview !== undefined ? formatValue('preview', record.preview, locale) : null,
    ].filter(Boolean).join(' ')
  }
  if ('title' in record || 'status' in record) {
    return [
      record.title !== undefined ? formatValue('title', record.title, locale) : null,
      record.url !== undefined ? formatValue('url', record.url, locale) : null,
      record.status !== undefined ? formatValue('status', record.status, locale) : null,
    ].filter(Boolean).join(' ')
  }
  return formatSafeToolPreview(record, locale)
}

function labelForKey(key: string, locale: Locale) {
  if (locale === 'zh') return keyLabels.zh[key as keyof typeof keyLabels.zh] ?? key
  return key
}

export function formatSafeToolPreview(value: Record<string, unknown> | null | undefined, locale: Locale = 'en') {
  if (!value) return ''
  const webSearchPreview = formatWebSearchPreview(value, locale)
  if (webSearchPreview) return webSearchPreview
  const text = Object.entries(value)
    .filter(([key]) => allowedKeys.has(key) || isSensitiveKey(key))
    .map(([key, item]) => `${labelForKey(key, locale)}: ${formatValue(key, item, locale)}`)
    .filter(Boolean)
    .join(' · ')
  return text.length > 360 ? `${text.slice(0, 357)}...` : text
}

function recordValue(value: unknown) {
  return typeof value === 'object' && value !== null && !Array.isArray(value) ? value as Record<string, unknown> : null
}

export function buildToolEventPreview(event: RunEvent, loopCopy = '', locale: Locale = 'en') {
  const toolName = typeof event.metadata?.tool_name === 'string' ? event.metadata.tool_name : undefined
  const primary = `${humanToolName(toolName, locale)} ${statusCopy(event, locale)}`
  const details = [
    loopCopy,
    formatSafeToolPreview(recordValue(event.metadata?.arguments_summary), locale),
    formatSafeToolPreview(recordValue(event.metadata?.result_summary), locale),
  ].filter(Boolean).join(' · ')
  return { primary, details }
}
