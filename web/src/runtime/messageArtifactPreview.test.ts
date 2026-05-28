import { describe, expect, test } from 'bun:test'
import { extractMessageArtifact, stripMessageArtifactSource } from './messageArtifactPreview'

describe('messageArtifactPreview', () => {
  test('extracts fenced md blocks from assistant text as preview artifacts', () => {
    const message = {
      id: 'msg-a',
      content: '把下面内容保存为 `三句话.md`：\n\n```md\n# 三句话的 Markdown\n\n今天我开始写一个简单的 Markdown 文档。\n```',
    }

    expect(extractMessageArtifact(message)).toMatchObject({
      id: 'message:msg-a:markdown',
      title: '三句话的 Markdown',
      filename: '三句话.md',
      kind: 'markdown',
      content: '# 三句话的 Markdown\n\n今天我开始写一个简单的 Markdown 文档。',
    })
    expect(stripMessageArtifactSource(message.content)).not.toContain('```md')
    expect(stripMessageArtifactSource(message.content)).not.toContain('# 三句话的 Markdown')
  })

  test('extracts accidental inline md heading payloads', () => {
    const message = {
      id: 'msg-b',
      content: '把下面内容保存为 三句话.md： `md#三句话的 Markdown 今天我开始写一个简单的 Markdown 文档。`',
    }

    expect(extractMessageArtifact(message)).toMatchObject({
      title: '三句话的 Markdown 今天我开始写一个简单的 Markdown 文档。',
      filename: '三句话.md',
      content: '# 三句话的 Markdown 今天我开始写一个简单的 Markdown 文档。',
    })
    expect(stripMessageArtifactSource(message.content)).not.toContain('md#')
  })

  test('does not turn ordinary markdown answers into document cards', () => {
    const message = {
      id: 'msg-ordinary',
      content: 'markdown#文档标题##简介这里写一段简短介绍。',
    }

    expect(extractMessageArtifact(message)).toBeNull()
    expect(stripMessageArtifactSource(message.content)).toBe(message.content)
  })

  test('extracts artifact protocol links without treating the link text as source content', () => {
    const message = {
      id: 'msg-c',
      content: '已创建 [三句话的 Markdown](artifact:art_123)，你可以在右侧预览。',
    }

    expect(extractMessageArtifact(message)).toMatchObject({
      id: 'art_123',
      title: '三句话的 Markdown',
      filename: '三句话的 Markdown',
      kind: 'markdown',
    })
    expect(stripMessageArtifactSource(message.content)).toBe('已创建 三句话的 Markdown，你可以在右侧预览。')
  })
})
