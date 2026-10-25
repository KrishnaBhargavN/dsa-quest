import type { ReactNode } from 'react'

// A deliberately small Markdown renderer for the theory lessons. It builds
// React elements (never raw HTML) and supports only what the lessons use:
// ## and ### headings, paragraphs, - and 1. lists, | tables |, fenced code,
// `inline code` and **bold**. Java code blocks start collapsed so a cold
// attempt isn't spoiled.

function inline(text: string): ReactNode[] {
  return text.split(/(`[^`]+`|\*\*[^*]+\*\*)/g).map((part, i) => {
    if (part.startsWith('`') && part.endsWith('`') && part.length > 1) return <code key={i}>{part.slice(1, -1)}</code>
    if (part.startsWith('**') && part.endsWith('**') && part.length > 3) return <strong key={i}>{part.slice(2, -2)}</strong>
    return part
  })
}

function cells(row: string) {
  return row.trim().replace(/^\||\|$/g, '').split('|').map(c => c.trim())
}

export function Markdown({ source }: { source: string }) {
  const lines = source.split('\n')
  const out: ReactNode[] = []
  let i = 0
  while (i < lines.length) {
    const line = lines[i]
    const key = out.length

    if (line.startsWith('```')) {
      const lang = line.slice(3).trim()
      const body: string[] = []
      i++
      while (i < lines.length && !lines[i].startsWith('```')) body.push(lines[i++])
      i++ // closing fence
      const pre = <pre className="md-code"><code>{body.join('\n')}</code></pre>
      out.push(lang === 'java'
        ? <details key={key} className="md-details"><summary>Show Java code</summary>{pre}</details>
        : <div key={key}>{pre}</div>)
      continue
    }
    if (line.startsWith('### ')) { out.push(<h4 key={key} className="md-h4">{inline(line.slice(4))}</h4>); i++; continue }
    if (line.startsWith('## ')) { out.push(<h3 key={key} className="md-h3">{inline(line.slice(3))}</h3>); i++; continue }

    if (line.startsWith('|')) {
      const rows: string[] = []
      while (i < lines.length && lines[i].startsWith('|')) rows.push(lines[i++])
      const [head, , ...body] = rows
      out.push(
        <div key={key} className="md-table-wrap">
          <table className="md-table">
            <thead><tr>{cells(head).map((c, j) => <th key={j}>{inline(c)}</th>)}</tr></thead>
            <tbody>{body.map((r, j) => <tr key={j}>{cells(r).map((c, k) => <td key={k}>{inline(c)}</td>)}</tr>)}</tbody>
          </table>
        </div>,
      )
      continue
    }

    const bullet = /^- /, numbered = /^\d+\. /
    if (bullet.test(line) || numbered.test(line)) {
      const re = bullet.test(line) ? bullet : numbered
      const items: string[] = []
      while (i < lines.length && re.test(lines[i])) items.push(lines[i++].replace(re, ''))
      const lis = items.map((t, j) => <li key={j}>{inline(t)}</li>)
      out.push(re === bullet ? <ul key={key} className="md-list">{lis}</ul> : <ol key={key} className="md-list">{lis}</ol>)
      continue
    }

    if (line.trim() === '') { i++; continue }

    const para: string[] = []
    while (i < lines.length && lines[i].trim() !== '' && !/^(```|#{2,3} |\||- |\d+\. )/.test(lines[i])) para.push(lines[i++])
    out.push(<p key={key}>{inline(para.join(' '))}</p>)
  }
  return <div className="md">{out}</div>
}
