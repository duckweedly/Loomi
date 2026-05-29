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

  test('extracts bare generated markdown documents as preview artifacts', () => {
    const message = {
      id: 'msg-doc',
      content: 'markdown#项目名称一句话介绍这个项目是做什么的。##目录-[项目简介](#项目简介)-[功能特性](#功能特性)-[快速开始](#快速开始)##项目简介这里填写项目背景。##功能特性-支持核心功能一-支持核心功能二##许可证MIT',
    }

    expect(extractMessageArtifact(message)).toMatchObject({
      id: 'message:msg-doc:markdown',
      title: '项目名称',
      filename: 'Markdown.md',
      kind: 'markdown',
    })
    expect(extractMessageArtifact(message)?.content).toContain('# 项目名称')
    expect(extractMessageArtifact(message)?.content).toContain('## 目录')
    expect(stripMessageArtifactSource(message.content)).toBe('')
  })

  test('extracts inline markdown file content prompts without explicit save intent', () => {
    const message = {
      id: 'msg-file-content',
      content: '如果你想要“一个 Markdown 文件内容”，直接用下面这个通用版： ` markdown #文档标题##概述这里写文档的简要说明。 ##目标-目标一-目标二##内容###1. 第一部分这里写第一部分内容。 ` 如果你要特定类型，告诉我类型即可。',
    }

    expect(extractMessageArtifact(message)).toMatchObject({
      id: 'message:msg-file-content:markdown',
      title: '文档标题',
      filename: 'Markdown.md',
      kind: 'markdown',
      content: '# 文档标题\n\n## 概述\n\n这里写文档的简要说明。\n\n## 目标\n- 目标一\n- 目标二\n\n## 内容\n\n1. 第一部分这里写第一部分内容。',
    })
    expect(stripMessageArtifactSource(message.content)).toBe('如果你想要“一个 Markdown 文件内容”，直接用下面这个通用版：  如果你要特定类型，告诉我类型即可。')
  })

  test('extracts loose markdown file payloads when stream recovery drops code fences', () => {
    const message = {
      id: 'msg-loose-file-content',
      content: '如果你想要“一个Markdown文件内容”，直接用下面这个通用版：markdown #文档标题##概述这里写文档的简要说明。##目标-目标一-目标二如果你要特定类型，告诉我类型即可。',
    }

    expect(extractMessageArtifact(message)).toMatchObject({
      id: 'message:msg-loose-file-content:markdown',
      title: '文档标题',
      content: '# 文档标题\n\n## 概述\n\n这里写文档的简要说明。\n\n## 目标\n- 目标一\n- 目标二',
    })
    expect(stripMessageArtifactSource(message.content)).toBe('如果你想要“一个Markdown文件内容”，直接用下面这个通用版：如果你要特定类型，告诉我类型即可。')
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

  test('resolves artifact protocol links from known artifact metadata', () => {
    const message = {
      id: 'msg-svg-link',
      content: '已生成 [LangGraph SVG 讲解图](artifact:art_svg)。',
    }

    expect(extractMessageArtifact(message, [{
      id: 'art_svg',
      title: 'LangGraph SVG 讲解图',
      filename: 'LangGraph SVG 讲解图.svg',
      mimeType: 'image/svg+xml',
      kind: 'svg',
      content: '<svg viewBox="0 0 10 10"></svg>',
    }])).toMatchObject({
      id: 'art_svg',
      kind: 'svg',
      mimeType: 'image/svg+xml',
      content: '<svg viewBox="0 0 10 10"></svg>',
    })
  })

  test('extracts raw SVG answers as visual artifacts instead of chat text', () => {
    const message = {
      id: 'msg-svg',
      content: '下面是图：\n```svg\n<svg viewBox="0 0 20 20"><title>流程图</title><rect width="20" height="20"/></svg>\n```',
    }

    expect(extractMessageArtifact(message)).toMatchObject({
      id: 'message:msg-svg:svg',
      title: '流程图',
      filename: 'visual.svg',
      mimeType: 'image/svg+xml',
      kind: 'svg',
    })
    expect(stripMessageArtifactSource(message.content)).toBe('下面是图：')
  })
})
