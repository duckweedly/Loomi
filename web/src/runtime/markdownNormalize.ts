const fencedCodeBlockPattern = /(```[\s\S]*?```)/g

export function normalizeStreamingFenceStart(content: string) {
  return content
    .replace(/```(sql)(?=(?:CREATE|SELECT|WITH|INSERT|UPDATE|DELETE|ALTER|DROP|TRUNCATE|MERGE|EXPLAIN)\b)/gi, '```$1\n')
    .replace(/```(tsx|ts|jsx|js|javascript|typescript)(?=(?:import|export|const|let|var|function|class|interface|type)\b)/gi, '```$1\n')
    .replace(/```(python|py)(?=(?:from|import|def|class|if|for|while|with|print)\b)/gi, '```$1\n')
    .replace(/```(json)(?=[[{])/gi, '```$1\n')
    .replace(/```(bash|sh|zsh)(?=(?:cd|ls|cat|grep|rg|npm|pnpm|bun|yarn|git|curl|echo)\b)/gi, '```$1\n')
}

function normalizeMarkdownProse(content: string) {
  return content
    .replace(/(^|\n)(?:md|markdown)(?=#{1,6}\S)/gi, '$1')
    .replace(/([^#(\n])(?=#{1,6}(?:[一二三四五六七八九十]+、|\d+[.、]|\S+\s))/g, '$1\n\n')
    .replace(/\s*---\s*(#{1,6})/g, '\n\n$1')
    .replace(/^[\t\u00a0\u3000 ]{1,3}\\(#{1,6})/gm, '$1')
    .replace(/^[\t\u00a0\u3000 ]{1,3}(#{1,6})/gm, '$1')
    .replace(/^\\(#{1,6})(?=\s)/gm, '$1')
    .replace(/([^\n])\s+(#{1,6})(?=\S)/g, '$1\n\n$2')
    .replace(/^(#{1,6}[^\n-]{1,40})-(?=\[)/gm, '$1\n- ')
    .replace(/^(#{1,6}[^\n-]{1,20})-(?=[\p{Script=Han}A-Za-z0-9])/gmu, '$1\n- ')
    .replace(/^(#{1,6})(?!#)(?=\S)/gm, '$1 ')
    .replace(/([)\]])-(?=\[)/g, '$1\n- ')
    .replace(/^(#{1,6}\s+\d+)\.(?=\S)/gm, '$1. ')
    .replace(/^(#{1,6}\s+[一二三四五六七八九十]+、[^\n。！？：:]{6,80}?)(到\s*\d{4}\s*年|目前|现在|整体|主流)/gm, '$1\n\n$2')
    .replace(/([：:。；;])\s+(\d+)\.-/g, '$1\n\n$2. ')
    .replace(/([。；;])\s+(\d+)\.-/g, '$1\n\n$2. ')
    .replace(/\s+(\d+)\.-/g, '\n$1. ')
    .replace(/([。；;])\s+-(?=\S)/g, '$1\n- ')
    .replace(/(?<![|\n])\s+-(?=\S)/g, '\n- ')
    .replace(/([^\n])-(?=[\p{Script=Han}])/gmu, '$1\n- ')
}

export function normalizeMarkdownContent(content: string) {
  return normalizeStreamingFenceStart(content.replace(/\r\n/g, '\n'))
    .split(fencedCodeBlockPattern)
    .map((part, index) => (index % 2 === 1 ? part : normalizeMarkdownProse(part)))
    .join('')
}
