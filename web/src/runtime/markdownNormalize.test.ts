import { describe, expect, test } from 'bun:test'
import { normalizeMarkdownContent, normalizeStreamingFenceStart } from './markdownNormalize'

describe('markdownNormalize', () => {
  test('adds missing line break after streamed code fence language labels', () => {
    expect(normalizeStreamingFenceStart('```sqlSELECT * FROM runs;')).toBe('```sql\nSELECT * FROM runs;')
    expect(normalizeStreamingFenceStart('```tsximport { useState } from "react"')).toBe('```tsx\nimport { useState } from "react"')
  })

  test('repairs dense Chinese report headings and list markers', () => {
    const normalized = normalizeMarkdownContent('2026 年 5 月值得关注的 AIAgent 开源项目##一、 2026 年 AIAgent 开源生态的几个核心变化到 2026 年，主流方向明显变成： 1.-状态管理。 -代表： LangGraph。 2.-工具协议。##二、 重点开源项目')

    expect(normalized).toContain('项目\n\n## 一、 2026 年 AIAgent 开源生态的几个核心变化\n\n到 2026 年')
    expect(normalized).toContain('变成：\n\n1. 状态管理。')
    expect(normalized).toContain('\n- 代表： LangGraph。')
    expect(normalized).toContain('\n2. 工具协议。')
    expect(normalized).toContain('\n\n## 二、 重点开源项目')
  })

  test('removes accidental md prefix before markdown headings', () => {
    expect(normalizeMarkdownContent('md#三句话的 Markdown\n正文')).toContain('# 三句话的 Markdown')
  })

  test('renders compact markdown language-prefixed documents as markdown prose', () => {
    const normalized = normalizeMarkdownContent('markdown#文档标题##简介这里写一段简短介绍。##目录-[功能](#功能)-[使用方法](#使用方法)##功能-功能一-功能二##使用方法 1.第一步 2.第二步')

    expect(normalized).toContain('# 文档标题')
    expect(normalized).toContain('\n\n## 简介')
    expect(normalized).toContain('\n\n## 目录\n- [功能](#功能)')
    expect(normalized).toContain('\n\n## 功能\n- 功能一')
    expect(normalized).not.toContain('markdown#')
  })

  test('repairs dense README-like markdown with no spaces after headings', () => {
    const normalized = normalizeMarkdownContent('markdown#项目名称>简短描述：这个项目用于解决什么问题，适合什么场景。###项目介绍这里填写项目的背景、目标和主要用途。##主要功能-功能一：说明功能用途-功能二：说明功能用途-支持快速部署###技术栈-前端：React/Vue/HTML-后端：Node.js/Python/Java##快速开始#### 1. 克隆项目#### 2. 安装依赖### 3. 启动开发环境## 使用说明打开浏览器访问：## 项目结构## 环境变量## 常用命令')

    expect(normalized).toContain('# 项目名称\n\n> 简短描述')
    expect(normalized).toContain('\n\n### 项目介绍\n\n这里填写项目的背景')
    expect(normalized).toContain('\n\n## 主要功能\n- 功能一')
    expect(normalized).toContain('\n\n### 技术栈\n- 前端：React/Vue/HTML')
    expect(normalized).toContain('\n\n## 快速开始')
    expect(normalized).toContain('\n1. 克隆项目')
    expect(normalized).toContain('\n2. 安装依赖')
    expect(normalized).toContain('\n3. 启动开发环境')
    expect(normalized).toContain('\n\n## 使用说明\n\n打开浏览器访问：')
    expect(normalized).not.toContain('###项目介绍')
    expect(normalized).not.toContain('##快速开始')
  })

  test('repairs collapsed pipe tables from dense assistant summaries', () => {
    const normalized = normalizeMarkdownContent('##按文件类型分类|类型|数量||---|---:||Markdown文档|35||Excel表格|2||图片|1|##按大类分类|大类|数量||---|---:||文档|37||图片|2|')

    expect(normalized).toContain('## 按文件类型分类\n|类型|数量|\n|---|---:|\n|Markdown文档|35|')
    expect(normalized).toContain('\n\n## 按大类分类\n|大类|数量|\n|---|---:|\n|文档|37|')
    expect(normalized).not.toContain('||---')
  })

  test('unwraps whole markdown fences so generated documents render as prose', () => {
    const normalized = normalizeMarkdownContent('```markdown\n# 项目名称\n\n一句话介绍。\n\n## 目录\n- [功能](#功能)\n```')

    expect(normalized).toContain('# 项目名称')
    expect(normalized).toContain('\n\n## 目录\n- [功能](#功能)')
    expect(normalized).not.toContain('```markdown')
  })

  test('unwraps markdown envelopes without breaking inner code fences', () => {
    const normalized = normalizeMarkdownContent('```markdown\n# 标题\n\n正文。\n\n```bash\necho "Hello"\n```\n```')

    expect(normalized).toContain('# 标题')
    expect(normalized).toContain('```bash\necho "Hello"\n```')
    expect(normalized).not.toContain('```markdown')
  })

  test('drops empty fenced blocks that would render as blank code cards', () => {
    const normalized = normalizeMarkdownContent('## 许可证\n\nMIT\n\n```\n```')

    expect(normalized).toContain('## 许可证')
    expect(normalized).toContain('MIT')
    expect(normalized).not.toContain('```')
  })

  test('keeps real fenced code block closing fences', () => {
    const normalized = normalizeMarkdownContent('```bash\necho "Hello"\n```')

    expect(normalized).toBe('```bash\necho "Hello"\n```')
  })

  test('does not repair markdown-looking text inside fenced code blocks', () => {
    const normalized = normalizeMarkdownContent('正文---##标题\n```text\n---##not-heading\n1.-not-list\n```')

    expect(normalized).toContain('正文\n\n## 标题')
    expect(normalized).toContain('```text\n---##not-heading\n1.-not-list\n```')
  })
})
