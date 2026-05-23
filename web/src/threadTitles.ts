import type { Thread } from './domain'

const baseTitle = 'New thread'

export function createNextThreadTitle(threads: Pick<Thread, 'title'>[]) {
  const usedNumbers = new Set<number>()

  for (const thread of threads) {
    if (thread.title === baseTitle) {
      usedNumbers.add(1)
      continue
    }

    const match = /^New thread (\d+)$/.exec(thread.title)
    if (match) usedNumbers.add(Number(match[1]))
  }

  if (!usedNumbers.has(1)) return baseTitle

  let nextNumber = 2
  while (usedNumbers.has(nextNumber)) nextNumber += 1
  return `${baseTitle} ${nextNumber}`
}
