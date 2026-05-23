import { describe, expect, test } from 'bun:test'
import { existsSync, readFileSync, statSync } from 'node:fs'
import { resolve } from 'node:path'

describe('AgentStateMotion hedgehog asset', () => {
  test('uses the original Loomi hedgehog image from the standalone state prototype', () => {
    const source = readFileSync(resolve(import.meta.dir, 'AgentStateMotion.tsx'), 'utf8')
    const css = readFileSync(resolve(import.meta.dir, '../styles.css'), 'utf8')
    const assetPath = resolve(import.meta.dir, '../assets/loomi-hedgehog.png')
    const assetModule = readFileSync(resolve(import.meta.dir, '../assets/loomiHedgehog.ts'), 'utf8')

    expect(existsSync(assetPath)).toBe(true)
    expect(statSync(assetPath).size).toBeGreaterThan(400000)
    expect(assetModule).toContain("import loomiHedgehogUrl from './loomi-hedgehog.png'")
    expect(source).toContain("import { loomiHedgehogImage } from '../assets/loomiHedgehog'")
    expect(source).toContain('src={loomiHedgehogImage}')
    expect(css).not.toContain('radial-gradient(circle at 31% 41%')
  })
})
