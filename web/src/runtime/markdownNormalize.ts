const fencedCodeBlockPattern = /(```[\s\S]*?```)/g
const emptyFencedCodeBlockPattern = /(^|\n)[\t ]*```[\w-]*[\t ]*\n[\t ]*```[\t ]*(?=\n|$)/g

export function normalizeStreamingFenceStart(content: string) {
  return content
    .replace(/```(markdown|md)(?=#{1,6}\S)/gi, '```$1\n')
    .replace(/```(sql)(?=(?:CREATE|SELECT|WITH|INSERT|UPDATE|DELETE|ALTER|DROP|TRUNCATE|MERGE|EXPLAIN)\b)/gi, '```$1\n')
    .replace(/```(tsx|ts|jsx|js|javascript|typescript)(?=(?:import|export|const|let|var|function|class|interface|type)\b)/gi, '```$1\n')
    .replace(/```(python|py)(?=(?:from|import|def|class|if|for|while|with|print)\b)/gi, '```$1\n')
    .replace(/```(json)(?=[[{])/gi, '```$1\n')
    .replace(/```(bash|sh|zsh)(?=(?:cd|ls|cat|grep|rg|npm|pnpm|bun|yarn|git|curl|echo)\b)/gi, '```$1\n')
}

function normalizeMarkdownProse(content: string) {
  const normalized = content
    .replace(/(^|\n)(?:md|markdown)(?=#{1,6}\S)/gi, '$1')
    .replace(/([^#(\n])(?=#{1,6}(?:[一二三四五六七八九十]+、|\d+[.、]|\S+\s))/g, '$1\n\n')
    .replace(/([^#(\n])(?=#{1,6}[\p{Script=Han}A-Za-z0-9])/gmu, '$1\n\n')
    .replace(/\s*---\s*(#{1,6})/g, '\n\n$1')
    .replace(/^[\t\u00a0\u3000 ]{1,3}\\(#{1,6})/gm, '$1')
    .replace(/^[\t\u00a0\u3000 ]{1,3}(#{1,6})/gm, '$1')
    .replace(/^\\(#{1,6})(?=\s)/gm, '$1')
    .replace(/([^\n])\s+(#{1,6})(?=\S)/g, '$1\n\n$2')
    .replace(/^(#\s+[\p{Script=Han}A-Za-z0-9 _-]{2,30}?)(一句话|简短描述|概述|简介|这里)/gmu, '$1\n\n$2')
    .replace(/^(#{1,6})(概述|目标|目录|内容|待办事项|备注|安装依赖|环境要求|运行访问|启动项目|构建|目录结构|配置说明|示例代码|更新日志|项目介绍|项目简介|主要功能|功能特性|技术栈|快速开始|使用方法|使用说明|项目结构|环境变量|常用命令|贡献指南|许可证|License|安装|配置|贡献)(?=\S)/gmu, '$1 $2\n\n')
    .replace(/^(#{1,6}\s+)(概述|目标|目录|内容|待办事项|备注|安装依赖|环境要求|运行访问|启动项目|构建|目录结构|配置说明|示例代码|更新日志|项目介绍|项目简介|主要功能|功能特性|技术栈|快速开始|使用方法|使用说明|项目结构|环境变量|常用命令|贡献指南|许可证|License|安装|配置|贡献)(?=\S)/gmu, '$1$2\n\n')
    .replace(/^(#{1,6}[^\n-]{1,40})-(?=\[)/gm, '$1\n- ')
    .replace(/^(#{1,6}[^\n-]{1,20})-(?=[\p{Script=Han}A-Za-z0-9])/gmu, '$1\n- ')
    .replace(/^(#{1,6})(?!#)(?=\S)/gm, '$1 ')
    .replace(/^(#{1,6}\s+[^\n>]{1,36})>(?=\S)/gm, '$1\n\n> ')
    .replace(/^#{3,6}\s+(\d+)\.\s*/gm, '$1. ')
    .replace(/([)\]])-(?=\[)/g, '$1\n- ')
    .replace(/^(#{1,6}\s+\d+)\.(?=\S)/gm, '$1. ')
    .replace(/^(#{1,6}\s+[一二三四五六七八九十]+、[^\n。！？：:]{6,80}?)(到\s*\d{4}\s*年|目前|现在|整体|主流)/gm, '$1\n\n$2')
    .replace(/([：:。；;])\s+(\d+)\.-/g, '$1\n\n$2. ')
    .replace(/([。；;])\s+(\d+)\.-/g, '$1\n\n$2. ')
    .replace(/\s+(\d+)\.-/g, '\n$1. ')
    .replace(/([。；;])\s+-(?=\S)/g, '$1\n- ')
    .replace(/(?<![|\n])\s+-(?=\S)/g, '\n- ')
    .replace(/([^\s|\n])-(?=(?:npm|pnpm|yarn|Node\.js|Python|Docker|前端|后端|数据库|其他)\b)/g, '$1\n- ')
    .replace(/([^\n])-(?=[\p{Script=Han}])/gmu, '$1\n- ')
  return normalizeCollapsedPipeTables(normalized)
}

function normalizeCollapsedPipeTables(content: string) {
  return content
    .split('\n')
    .flatMap((line) => {
      if (!/\|\|[-: ]{0,3}-{3,}/.test(line)) return [line]
      const withHeaderBreak = line.replace(/([^|\n])(\|[^|\n]+(?:\|[^|\n]+)+\|\|[-: ]{0,3}-{3,})/, '$1\n$2')
      return withHeaderBreak.split('\n').map((part) => (
        /\|\|[-: ]{0,3}-{3,}/.test(part) ? part.replace(/\|\|/g, '|\n|') : part
      ))
    })
    .join('\n')
}

function unwrapMarkdownEnvelope(content: string) {
  const opening = content.match(/^\s*```(?:markdown|md)[\t ]*\n/i)
  if (!opening) return content

  let body = content.slice(opening[0].length)
  const trimmed = body.trimEnd()
  if (trimmed.endsWith('```')) {
    const beforeTrailingFence = trimmed.slice(0, -3)
    if (!beforeTrailingFence.includes('```') || beforeTrailingFence.trimEnd().endsWith('```')) {
      body = beforeTrailingFence
    }
  }

  return body.trim()
}

function dropEmptyFencedCodeBlocks(content: string) {
  return content.replace(emptyFencedCodeBlockPattern, '$1')
}

export function normalizeMarkdownContent(content: string) {
  return dropEmptyFencedCodeBlocks(unwrapMarkdownEnvelope(normalizeStreamingFenceStart(content.replace(/\r\n/g, '\n'))))
    .split(fencedCodeBlockPattern)
    .map((part, index) => (index % 2 === 1 ? part : normalizeMarkdownProse(part)))
    .join('')
}
