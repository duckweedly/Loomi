export function nextTypewriterFrame(from: string, to: string, revealChars: number): string {
  if (from === to || revealChars <= 0) return from
  let prefixLength = 0
  while (prefixLength < from.length && prefixLength < to.length && from[prefixLength] === to[prefixLength]) {
    prefixLength += 1
  }
  const nextSegment = to.slice(prefixLength)
  if (revealChars >= nextSegment.length * 2) return to
  const visibleLength = Math.max(1, Math.floor(revealChars / 2))
  return `${to.slice(0, prefixLength)}${nextSegment.slice(0, visibleLength)}${from.slice(prefixLength)}`
}
