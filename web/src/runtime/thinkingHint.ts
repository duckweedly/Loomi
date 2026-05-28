import type { Locale } from '../i18n'
import { getDictionary } from '../i18n'

const thinkingHintStoragePrefix = 'loomi:run-thinking-hint:'

function chooseThinkingHint(hints: readonly string[]) {
  if (hints.length === 0) return ''
  return hints[Math.floor(Math.random() * hints.length)] ?? hints[0] ?? ''
}

function stableThinkingHint(runId: string, hints: readonly string[]) {
  let hash = 0
  for (let index = 0; index < runId.length; index += 1) {
    hash = ((hash << 5) - hash + runId.charCodeAt(index)) | 0
  }
  return hints[Math.abs(hash) % hints.length] ?? hints[0] ?? ''
}

export function thinkingHintForRun(runId: string, locale: Locale) {
  const hints = getDictionary(locale).chatCanvas.thinkingHints
  if (hints.length === 0) return getDictionary(locale).chatCanvas.modelDrafting
  if (typeof window === 'undefined' || !window.localStorage) return stableThinkingHint(runId, hints)
  const key = `${thinkingHintStoragePrefix}${runId}`
  const stored = window.localStorage.getItem(key)
  if (stored && hints.includes(stored)) return stored
  const hint = chooseThinkingHint(hints) || stableThinkingHint(runId, hints)
  window.localStorage.setItem(key, hint)
  return hint
}

export function thinkingHintWithElapsed(runId: string, locale: Locale, startedAt?: string) {
  const started = startedAt ? new Date(startedAt).getTime() : Date.now()
  const elapsed = Number.isNaN(started) ? 0 : Math.max(0, Math.round((Date.now() - started) / 1000))
  return `${thinkingHintForRun(runId, locale)} ${elapsed}s`
}
