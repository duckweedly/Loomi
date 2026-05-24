import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

describe('ChatCanvas state copy', () => {
  test('gets sparse Chinese labels from the i18n dictionary', () => {
    const source = readFileSync(resolve(import.meta.dir, '../i18n.ts'), 'utf8')

    expect(source).toContain('未选择会话')
    expect(source).toContain('新对话')
    expect(source).toContain('加载中')
    expect(source).toContain('加载失败')
    expect(source).toContain('等待执行')
    expect(source).toContain('执行中')
    expect(source).toContain('已完成')
    expect(source).toContain('执行失败')
    expect(source).toContain('后端能力未接入')
    expect(source).toContain('模型网关')
    expect(source).toContain('工具调用未执行')
    expect(source).toContain('已停止生成')
  })

  test('routes ChatCanvas rendering through deriveChatCanvasState', () => {
    const source = readFileSync(resolve(import.meta.dir, 'ChatCanvas.tsx'), 'utf8')

    expect(source).toContain("from '../runtime/chatCanvasState'")
    expect(source).toContain('deriveChatCanvasState')
  })
})
