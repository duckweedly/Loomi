import type { Thread } from './domain'

export function filterThreadsByMode(threads: Thread[], mode: Thread['mode']) {
  return threads.filter((thread) => thread.mode === mode)
}
