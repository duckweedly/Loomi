import type { PreviewArtifact } from '../runtime/artifactPreview'

function escapeHTML(value: string) {
  return value.replaceAll('&', '&amp;').replaceAll('<', '&lt;').replaceAll('>', '&gt;')
}

function frameDocument(artifact: PreviewArtifact) {
  const content = artifact.content || artifact.excerpt || ''
  const csp = "default-src 'none'; img-src data: blob: https:; style-src 'unsafe-inline'; script-src 'unsafe-inline'; font-src data: https:;"
  const body = artifact.kind === 'svg'
    ? `<main class="loomi-svg-stage">${content}</main>`
    : content
  return `<!doctype html>
<html>
<head>
<meta charset="utf-8" />
<meta http-equiv="Content-Security-Policy" content="${csp}" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<style>
html,body{margin:0;min-height:100%;background:transparent;color:#111827;font-family:Inter,ui-sans-serif,system-ui,sans-serif}
body{display:grid;place-items:start center;padding:16px;box-sizing:border-box}
.loomi-svg-stage{width:100%;display:grid;place-items:center}
.loomi-svg-stage svg,svg{max-width:100%;height:auto}
</style>
</head>
<body>${body || `<pre>${escapeHTML(artifact.mimeType)}</pre>`}</body>
</html>`
}

export function ArtifactFrame({ artifact, title }: { artifact: PreviewArtifact; title?: string }) {
  return (
    <iframe
      className="artifact-frame"
      title={title ?? artifact.title}
      sandbox="allow-scripts"
      referrerPolicy="no-referrer"
      srcDoc={frameDocument(artifact)}
    />
  )
}
