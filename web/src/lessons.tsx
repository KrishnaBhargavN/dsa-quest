import { createContext, useContext, useEffect, useState } from 'react'
import { api, type Lesson, type LessonMeta } from './api'
import { Markdown } from './md'
import { errText } from './ui'

type LessonsCtx = { index: LessonMeta[]; open: (id: string) => void }

export const LessonsContext = createContext<LessonsCtx>({ index: [], open: () => {} })

export function useLessons() {
  return useContext(LessonsContext)
}

/** A one-line link to a lesson, e.g. on a problem card. */
export function LessonLink({ id, prefix = 'Theory' }: { id: string; prefix?: string }) {
  const { index, open } = useLessons()
  const meta = index.find(l => l.id === id)
  if (!meta) return null
  return (
    <button className="lesson-link" onClick={() => open(id)}>
      <span>{prefix}: {meta.title}</span>
      <span className="muted small">{meta.minutes} min read</span>
    </button>
  )
}

/** Full-screen reader. Closing goes back in history, so the phone's back button closes it too. */
export function LessonView({ id, close }: { id: string; close: () => void }) {
  const [lesson, setLesson] = useState<Lesson | null>(null)
  const [err, setErr] = useState('')

  useEffect(() => {
    let live = true
    setLesson(null)
    api.lesson(id).then(l => { if (live) setLesson(l) }).catch(e => { if (live) setErr(errText(e)) })
    window.scrollTo(0, 0)
    return () => { live = false }
  }, [id])

  return (
    <div className="lesson" role="dialog" aria-label={lesson?.title ?? 'Lesson'}>
      <header className="lesson-bar">
        <button className="btn small" onClick={close}>‹ Back</button>
        {lesson && <span className="muted small">{lesson.group} · {lesson.minutes} min</span>}
      </header>
      <article className="lesson-body">
        {err && <p className="error">{err}</p>}
        {!lesson && !err && <p className="muted">Loading</p>}
        {lesson && (
          <>
            <h2 className="lesson-title">{lesson.title}</h2>
            <p className="lesson-summary">{lesson.summary}</p>
            <Markdown source={lesson.body} />
          </>
        )}
      </article>
    </div>
  )
}

/** Every lesson, grouped, for the List tab. */
export function TheoryIndex() {
  const { index, open } = useLessons()
  if (index.length === 0) return null
  const groups: { name: string; items: LessonMeta[] }[] = []
  for (const l of index) {
    const g = groups.find(x => x.name === l.group)
    if (g) g.items.push(l)
    else groups.push({ name: l.group, items: [l] })
  }
  return (
    <details className="phase">
      <summary><span>Theory</span><span className="muted small">{index.length} lessons</span></summary>
      <div className="stack">
        {groups.map(g => (
          <div key={g.name} className="stack-sm">
            <h3 className="topic">{g.name}</h3>
            <ul className="picks">
              {g.items.map(l => (
                <li key={l.id}>
                  <button onClick={() => open(l.id)}>
                    <span>{l.title}</span>
                    <span className="muted small">{l.summary}</span>
                  </button>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
    </details>
  )
}
