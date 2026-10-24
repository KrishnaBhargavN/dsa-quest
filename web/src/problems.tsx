import { useCallback, useEffect, useState } from 'react'
import {
  api,
  type Difficulty, type Problem, type ProblemInput, type ProblemList, type Revision, type StartBody, type Status,
} from './api'
import { addDays, cap, DifficultyPicker, errText, fmtDay, isoDay, localToday, parseDay, type Toast } from './ui'

type ToastFn = (t: Toast) => void

const PHASES: Record<number, string> = {
  1: 'Phase 1 · Oct 26 – Nov 8',
  2: 'Phase 2 · Nov 9 – Nov 29',
  3: 'Phase 3 · Nov 30 – Dec 20',
}

const MARKS: Record<Status, { mark: string; label: string }> = {
  todo: { mark: '', label: 'Not done yet' },
  solved: { mark: '✓', label: 'Solved' },
  stuck: { mark: '!', label: 'Got stuck' },
  before: { mark: '–', label: 'Done before' },
}

/** Something a revision can be scheduled for. */
export type ReviseTarget = { problemId?: string; url: string; title: string; difficulty: Difficulty }

function startBodyFor(r: { problemId?: string; url: string; title: string; difficulty: Difficulty }): StartBody {
  return r.problemId ? { problemId: r.problemId } : { url: r.url, title: r.title, difficulty: r.difficulty }
}

// ---------------------------------------------------------------- one problem

export function ProblemHead({ p }: { p: Problem }) {
  const topic = p.topic || (p.blind ? 'Topic hidden' : '')
  return (
    <div className="stack-sm">
      <h2 className="problem-title">
        {p.url ? <a href={p.url} target="_blank" rel="noreferrer">{p.title}</a> : p.title}
      </h2>
      <p className="meta">
        <span className={`diff ${p.difficulty}`}>{cap(p.difficulty)}</span>
        {topic && <span>{topic}</span>}
        {p.url.includes('cses.fi') && <span>CSES</span>}
      </p>
    </div>
  )
}

function Hints({ hints }: { hints: string[] }) {
  const [shown, setShown] = useState(0)
  if (hints.length === 0) return null
  return (
    <div className="stack-sm">
      {hints.slice(0, shown).map((h, i) => (
        <p key={i} className="hint"><strong>Hint {i + 1}.</strong> {h}</p>
      ))}
      {shown < hints.length && (
        <button className="link" onClick={() => setShown(n => n + 1)}>
          {shown === 0 ? 'Stuck? Show hint 1' : `Show hint ${shown + 1}`}
        </button>
      )}
    </div>
  )
}

export function ProblemBody({ p }: { p: Problem }) {
  return (
    <div className="stack-sm">
      {p.description && <p>{p.description}</p>}
      {p.note && <p className="note">{p.note}</p>}
      <Hints key={p.id} hints={p.hints} />
    </div>
  )
}

// ---------------------------------------------------------------- solve tab

export function DueRevisions({ revs, busy, onStart }: { revs: Revision[]; busy: boolean; onStart: (b: StartBody) => void }) {
  const today = localToday()
  return (
    <section className="stack" aria-label="Revisions due">
      <h2 className="section-title">Due for revision</h2>
      <ul className="rows">
        {revs.map(r => (
          <li key={r.id} className="row">
            <span className="row-name">
              {r.url ? <a href={r.url} target="_blank" rel="noreferrer">{r.title}</a> : r.title}
              <span className="muted small block">
                {cap(r.difficulty)}, {r.due < today ? `overdue since ${fmtDay(r.due)}` : 'due today'}
              </span>
            </span>
            <span />
            <button className="btn small" disabled={busy} onClick={() => onStart(startBodyFor(r))}>Start</button>
          </li>
        ))}
      </ul>
    </section>
  )
}

export function NextUp({ next, busy, onStart, refresh, toast }: {
  next: Problem | null
  busy: boolean
  onStart: (b: StartBody) => void
  refresh: () => Promise<void>
  toast: ToastFn
}) {
  const [picked, setPicked] = useState<Problem | null>(null)
  const [others, setOthers] = useState<ProblemList | null>(null)
  const [showOthers, setShowOthers] = useState(false)
  const p = picked ?? next

  const toggleOthers = async () => {
    if (showOthers) { setShowOthers(false); return }
    setShowOthers(true)
    try {
      setOthers(await api.problems())
    } catch (e) {
      toast({ title: errText(e) })
    }
  }

  if (!p) {
    return (
      <section className="card stack">
        <p className="eyebrow">Next up</p>
        <h2 className="problem-title">The whole list is done</h2>
        <p className="muted">Work through your redo list, or add your own problems in the List tab.</p>
      </section>
    )
  }

  return (
    <section className="card stack" aria-label="Next problem">
      <p className="eyebrow">{picked ? 'Picked for now' : 'Next up'}</p>
      <ProblemHead p={p} />
      <ProblemBody p={p} />
      <div className="actions">
        <button className="btn primary" disabled={busy} onClick={() => onStart({ problemId: p.id })}>Start timer</button>
        <div className="btn-row">
          <button className="btn small" disabled={busy} onClick={async () => {
            try {
              await api.doneBefore(p.id, true)
              toast({ title: 'Marked as done before', lines: ['No points, and it leaves Next up.'] })
              setPicked(null)
              setOthers(null)
              setShowOthers(false)
              await refresh()
            } catch (e) {
              toast({ title: errText(e) })
            }
          }}>Done before</button>
          <button className="btn small" aria-expanded={showOthers} onClick={toggleOthers}>
            {showOthers ? 'Hide other problems' : 'Show other problems'}
          </button>
        </div>
        {picked && <button className="link" onClick={() => setPicked(null)}>Back to next up</button>}
      </div>
      {showOthers && (
        others ? <OtherProblems list={others} current={p.id} onPick={q => { setPicked(q); setShowOthers(false) }} />
          : <p className="muted small">Loading</p>
      )}
    </section>
  )
}

function OtherProblems({ list, current, onPick }: { list: ProblemList; current: string; onPick: (p: Problem) => void }) {
  const upcoming = list.problems.filter(p => p.status === 'todo' && p.id !== current).slice(0, 5)
  const redo = list.problems.filter(p => p.redo && p.id !== current)
  const group = (title: string, ps: Problem[]) => ps.length > 0 && (
    <div className="stack-sm">
      <p className="small muted">{title}</p>
      <ul className="picks">
        {ps.map(p => (
          <li key={p.id}>
            <button onClick={() => onPick(p)}>
              <span>{p.title}</span>
              <span className="muted small">{cap(p.difficulty)}{p.topic ? ` · ${p.topic}` : p.blind ? ' · topic hidden' : ''}</span>
            </button>
          </li>
        ))}
      </ul>
    </div>
  )
  return (
    <div className="stack">
      {group('Coming up next', upcoming)}
      {group('Your redo list', redo)}
      {upcoming.length === 0 && redo.length === 0 && <p className="muted small">Nothing else to pick right now.</p>}
    </div>
  )
}

export function RevisePrompt({ target, done, toast, inline }: {
  target: ReviseTarget
  done: (scheduled: boolean) => void
  toast: ToastFn
  inline?: boolean
}) {
  const [date, setDate] = useState('')
  const [busy, setBusy] = useState(false)
  const schedule = async (when: { inDays?: number; due?: string }) => {
    setBusy(true)
    try {
      const body = target.problemId
        ? { problemId: target.problemId, ...when }
        : { url: target.url, title: target.title, difficulty: target.difficulty, ...when }
      const r = await api.scheduleRevision(body)
      toast({ title: 'Revision scheduled', lines: [`${target.title}, ${fmtDay(r.due)}`] })
      done(true)
    } catch (e) {
      toast({ title: errText(e) })
    } finally {
      setBusy(false)
    }
  }
  return (
    <section className={inline ? 'stack panel' : 'card stack'} aria-label="Schedule a revision">
      <h2 className="section-title">Revise this later?</h2>
      <p className="muted small">
        {inline ? '' : `${target.title}. `}It shows up in the calendar and at the top of the Solve tab on that day.
      </p>
      <div className="btn-row">
        {[[3, 'In 3 days'], [7, 'In a week'], [14, 'In 2 weeks']].map(([d, label]) => (
          <button key={d} className="btn small" disabled={busy} onClick={() => schedule({ inDays: d as number })}>{label}</button>
        ))}
      </div>
      <div className="date-pick">
        <label className="field">
          <span>Or pick a date</span>
          <input type="date" min={addDays(localToday(), 1)} value={date} onChange={e => setDate(e.target.value)} />
        </label>
        <button className="btn small" disabled={busy || !date} onClick={() => schedule({ due: date })}>Schedule</button>
      </div>
      <button className="link" onClick={() => done(false)}>{inline ? 'Cancel' : 'Skip'}</button>
    </section>
  )
}

// ---------------------------------------------------------------- list tab

export function ListScreen({ startProblem, toast }: { startProblem: (b: StartBody) => Promise<void>; toast: ToastFn }) {
  const [list, setList] = useState<ProblemList | null>(null)
  const [version, setVersion] = useState(0)
  const [adding, setAdding] = useState(false)

  const load = useCallback(async () => {
    try {
      setList(await api.problems())
    } catch (e) {
      toast({ title: errText(e) })
    }
  }, [toast])

  useEffect(() => { void load() }, [load, version])
  const changed = () => setVersion(v => v + 1)

  if (!list) return <p className="muted">Loading the list</p>

  const custom = list.problems.filter(p => p.custom)
  const redo = list.problems.filter(p => p.redo)
  const rowProps = { nextId: list.nextId, startProblem, changed, toast }

  return (
    <div className="stack-lg">
      <Calendar version={version} startProblem={startProblem} changed={changed} toast={toast} />

      <section className="stack" aria-label="Progress">
        <div className="shop-head">
          <h2 className="section-title">Problem list</h2>
          <span className="pill">{list.done} / {list.total}</span>
        </div>
        <div className="meter" role="progressbar" aria-valuemin={0} aria-valuemax={list.total} aria-valuenow={list.done}>
          <div className="meter-fill" style={{ width: `${list.total ? (list.done / list.total) * 100 : 0}%` }} />
        </div>
        <p className="muted small">Dates are a guide. If a week runs over, just carry on in order.</p>
      </section>

      {redo.length > 0 && (
        <section className="stack">
          <h2 className="section-title">Redo list</h2>
          <p className="muted small">Problems you got stuck on, and mediums that took over 40 minutes. Solving one cleanly takes it off.</p>
          <ProblemRows problems={redo} {...rowProps} />
        </section>
      )}

      {[1, 2, 3].map(phase => {
        const ps = list.problems.filter(p => p.phase === phase && !p.custom)
        if (ps.length === 0) return null
        const done = ps.filter(p => p.status !== 'todo').length
        return (
          <details key={phase} className="phase" open={ps.some(p => p.id === list.nextId)}>
            <summary>
              <span>{PHASES[phase]}</span>
              <span className="muted small">{done} / {ps.length}</span>
            </summary>
            <TopicGroups problems={ps} {...rowProps} />
          </details>
        )
      })}

      {custom.length > 0 && (
        <details className="phase" open>
          <summary><span>Your problems</span><span className="muted small">{custom.filter(p => p.status !== 'todo').length} / {custom.length}</span></summary>
          <ProblemRows problems={custom} {...rowProps} />
        </details>
      )}

      {adding
        ? <ProblemEditor done={saved => { setAdding(false); if (saved) changed() }} toast={toast} />
        : <button className="link" onClick={() => setAdding(true)}>Add your own problem</button>}
    </div>
  )
}

type RowProps = {
  nextId: string
  startProblem: (b: StartBody) => Promise<void>
  changed: () => void
  toast: ToastFn
}

function TopicGroups({ problems, ...rest }: { problems: Problem[] } & RowProps) {
  const groups: { name: string; ps: Problem[] }[] = []
  for (const p of problems) {
    const name = p.blind ? 'Mixed, topic hidden until finished' : (p.topic || 'Other')
    const last = groups[groups.length - 1]
    if (last && last.name === name) last.ps.push(p)
    else groups.push({ name, ps: [p] })
  }
  return (
    <div className="stack">
      {groups.map((g, i) => (
        <div key={i} className="stack-sm">
          <h3 className="topic">{g.name}</h3>
          <ProblemRows problems={g.ps} {...rest} />
        </div>
      ))}
    </div>
  )
}

function ProblemRows({ problems, ...rest }: { problems: Problem[] } & RowProps) {
  return (
    <ul className="prows">
      {problems.map(p => <ProblemRow key={p.id} p={p} {...rest} />)}
    </ul>
  )
}

function ProblemRow({ p, nextId, startProblem, changed, toast }: { p: Problem } & RowProps) {
  const [open, setOpen] = useState(false)
  const [mode, setMode] = useState<'view' | 'edit' | 'revise'>('view')
  const [busy, setBusy] = useState(false)
  const m = MARKS[p.status]

  const act = async (fn: () => Promise<void>) => {
    setBusy(true)
    try {
      await fn()
    } catch (e) {
      toast({ title: errText(e) })
    } finally {
      setBusy(false)
    }
  }

  return (
    <li className="prow">
      <button className="prow-head" aria-expanded={open} onClick={() => setOpen(o => !o)}>
        <span className={`mark ${p.status}`} title={m.label} aria-label={m.label}>{m.mark}</span>
        <span className="prow-title">
          {p.title}
          {p.id === nextId && <span className="tag">Next</span>}
          {p.revisionDue && <span className="tag soft">Revise {fmtDay(p.revisionDue)}</span>}
        </span>
        <span className={`diff ${p.difficulty}`}>{cap(p.difficulty)}</span>
      </button>
      {open && mode === 'edit' && (
        <ProblemEditor existing={p} done={saved => { setMode('view'); if (saved) changed() }} toast={toast} />
      )}
      {open && mode === 'revise' && (
        <RevisePrompt inline target={{ problemId: p.id, url: p.url, title: p.title, difficulty: p.difficulty }}
          done={scheduled => { setMode('view'); if (scheduled) changed() }} toast={toast} />
      )}
      {open && mode === 'view' && (
        <div className="prow-body stack">
          {p.url && <a className="small" href={p.url} target="_blank" rel="noreferrer">Open the problem</a>}
          <ProblemBody p={p} />
          <div className="btn-row">
            <button className="btn small primary" disabled={busy} onClick={() => act(() => startProblem({ problemId: p.id }))}>Start timer</button>
            {(p.status === 'todo' || p.status === 'before') && (
              <button className="btn small" disabled={busy} onClick={() => act(async () => {
                await api.doneBefore(p.id, p.status === 'todo')
                changed()
              })}>{p.status === 'todo' ? 'Done before' : 'Undo done before'}</button>
            )}
            <button className="btn small" disabled={busy} onClick={() => setMode('revise')}>Revise later</button>
            {p.custom && <button className="btn small" disabled={busy} onClick={() => setMode('edit')}>Edit</button>}
          </div>
          <button className="link danger small" disabled={busy} onClick={() => act(async () => {
            if (!confirm(`Remove "${p.title}" from your list? Its history stays.`)) return
            await api.removeProblem(p.id)
            changed()
          })}>Remove from list</button>
        </div>
      )}
    </li>
  )
}

function ProblemEditor({ existing, done, toast }: { existing?: Problem; done: (saved: boolean) => void; toast: ToastFn }) {
  const [f, setF] = useState<ProblemInput>(() => ({
    title: existing?.title ?? '',
    url: existing?.url ?? '',
    difficulty: existing?.difficulty ?? 'medium',
    topic: existing?.topic ?? '',
    description: existing?.description ?? '',
    hints: [existing?.hints[0] ?? '', existing?.hints[1] ?? ''],
    placement: 'end',
  }))
  const [busy, setBusy] = useState(false)
  const set = (patch: Partial<ProblemInput>) => setF(x => ({ ...x, ...patch }))

  return (
    <form
      className="stack panel"
      onSubmit={async e => {
        e.preventDefault()
        if (!f.title.trim()) {
          toast({ title: 'Give it a title' })
          return
        }
        setBusy(true)
        try {
          if (existing) await api.updateProblem(existing.id, f)
          else await api.createProblem(f)
          toast({ title: existing ? 'Problem saved' : 'Problem added' })
          done(true)
        } catch (ex) {
          toast({ title: errText(ex) })
        } finally {
          setBusy(false)
        }
      }}
    >
      <h3 className="topic">{existing ? 'Edit problem' : 'Add your own problem'}</h3>
      <label className="field">
        <span>Title</span>
        <input value={f.title} onChange={e => set({ title: e.target.value })} />
      </label>
      <label className="field">
        <span>Link (optional)</span>
        <input type="url" inputMode="url" placeholder="https://leetcode.com/problems/…" value={f.url} onChange={e => set({ url: e.target.value })} />
      </label>
      <DifficultyPicker value={f.difficulty} onChange={d => set({ difficulty: d })} />
      <label className="field">
        <span>Topic (optional)</span>
        <input value={f.topic} onChange={e => set({ topic: e.target.value })} />
      </label>
      <label className="field">
        <span>Description (optional)</span>
        <textarea rows={2} value={f.description} onChange={e => set({ description: e.target.value })} />
      </label>
      {[0, 1].map(i => (
        <label key={i} className="field">
          <span>Hint {i + 1} (optional)</span>
          <input value={f.hints[i]} onChange={e => set({ hints: f.hints.map((h, j) => (j === i ? e.target.value : h)) })} />
        </label>
      ))}
      {!existing && (
        <div className="seg" role="radiogroup" aria-label="Where to put it">
          {(['next', 'end'] as const).map(pl => (
            <button key={pl} type="button" role="radio" aria-checked={f.placement === pl} className={f.placement === pl ? 'on' : ''} onClick={() => set({ placement: pl })}>
              {pl === 'next' ? 'Make it next up' : 'End of the list'}
            </button>
          ))}
        </div>
      )}
      <div className="btn-row">
        <button className="btn small primary" disabled={busy}>{existing ? 'Save' : 'Add problem'}</button>
        <button type="button" className="btn small" disabled={busy} onClick={() => done(false)}>Cancel</button>
      </div>
    </form>
  )
}

// ---------------------------------------------------------------- calendar

const WEEKDAYS = ['M', 'T', 'W', 'T', 'F', 'S', 'S']

function Calendar({ version, startProblem, changed, toast }: {
  version: number
  startProblem: (b: StartBody) => Promise<void>
  changed: () => void
  toast: ToastFn
}) {
  const [month, setMonth] = useState(() => localToday().slice(0, 7)) // YYYY-MM
  const [data, setData] = useState<{ today: string; revisions: Revision[] } | null>(null)
  const [selected, setSelected] = useState<string | null>(null)

  const first = month + '-01'
  const start = addDays(first, -((parseDay(first).getUTCDay() + 6) % 7)) // Monday on or before the 1st
  const end = addDays(start, 41)

  useEffect(() => {
    let live = true
    api.revisions(start, end)
      .then(d => { if (live) setData(d) })
      .catch(e => toast({ title: errText(e) }))
    return () => { live = false }
  }, [start, end, version, toast])

  const shiftMonth = (n: number) => {
    const d = parseDay(first)
    d.setUTCMonth(d.getUTCMonth() + n)
    setMonth(isoDay(d).slice(0, 7))
    setSelected(null)
  }

  const today = data?.today ?? localToday()
  const byDay = new Map<string, Revision[]>()
  for (const r of data?.revisions ?? []) byDay.set(r.due, [...(byDay.get(r.due) ?? []), r])
  const label = parseDay(first).toLocaleDateString(undefined, { month: 'long', year: 'numeric', timeZone: 'UTC' })
  const dayRevs = selected ? byDay.get(selected) ?? [] : []

  return (
    <section className="stack" aria-label="Revision calendar">
      <div className="shop-head">
        <h2 className="section-title">Revision calendar</h2>
        <div className="cal-nav">
          <button className="btn small" aria-label="Previous month" onClick={() => shiftMonth(-1)}>‹</button>
          <span className="cal-month">{label}</span>
          <button className="btn small" aria-label="Next month" onClick={() => shiftMonth(1)}>›</button>
        </div>
      </div>
      <div className="cal" role="grid" aria-label={label}>
        {WEEKDAYS.map((d, i) => <span key={i} className="cal-wd" aria-hidden="true">{d}</span>)}
        {Array.from({ length: 42 }, (_, i) => {
          const day = addDays(start, i)
          const revs = byDay.get(day) ?? []
          const open = revs.filter(r => !r.doneAt).length
          const cls = [
            'cal-day',
            day.slice(0, 7) !== month ? 'out' : '',
            day === today ? 'today' : '',
            day === selected ? 'sel' : '',
            open > 0 && day < today ? 'overdue' : '',
            revs.length > 0 && open === 0 ? 'cleared' : '',
          ].join(' ')
          return (
            <button key={day} className={cls} onClick={() => setSelected(day === selected ? null : day)}
              aria-label={`${fmtDay(day)}${revs.length ? `, ${revs.length} revision${revs.length > 1 ? 's' : ''}` : ''}`}>
              <span>{Number(day.slice(8))}</span>
              {revs.length > 0 && <span className="cal-count">{open > 0 ? open : '✓'}</span>}
            </button>
          )
        })}
      </div>
      {selected && (
        <div className="stack-sm">
          <p className="small"><strong>{fmtDay(selected)}</strong></p>
          {dayRevs.length === 0
            ? <p className="muted small">Nothing to revise. Schedule one from a problem in the list below, or after you finish one.</p>
            : <ul className="rows">{dayRevs.map(r => <RevisionRow key={r.id} r={r} startProblem={startProblem} changed={changed} toast={toast} />)}</ul>}
        </div>
      )}
    </section>
  )
}

function RevisionRow({ r, startProblem, changed, toast }: {
  r: Revision
  startProblem: (b: StartBody) => Promise<void>
  changed: () => void
  toast: ToastFn
}) {
  const [moving, setMoving] = useState(false)
  const [busy, setBusy] = useState(false)
  const act = async (fn: () => Promise<void>) => {
    setBusy(true)
    try {
      await fn()
      changed()
    } catch (e) {
      toast({ title: errText(e) })
    } finally {
      setBusy(false)
    }
  }
  return (
    <li className="rev">
      <span className="row-name">
        {r.url ? <a href={r.url} target="_blank" rel="noreferrer">{r.title}</a> : r.title}
        <span className="muted small block">{cap(r.difficulty)}{r.doneAt ? ', revised ✓' : ''}</span>
      </span>
      {!r.doneAt && (
        <div className="btn-row">
          <button className="btn small primary" disabled={busy} onClick={() => act(() => startProblem(startBodyFor(r)))}>Start</button>
          <button className="btn small" disabled={busy} onClick={() => setMoving(m => !m)}>Move</button>
          <button className="btn small" disabled={busy} onClick={() => act(() => api.revisionDone(r.id).then(() => undefined))}>Mark done</button>
          <button className="link danger small" disabled={busy} onClick={() => act(() => api.deleteRevision(r.id).then(() => undefined))}>Remove</button>
        </div>
      )}
      {moving && (
        <label className="field">
          <span>Move to</span>
          <input type="date" min={localToday()} defaultValue={r.due} onChange={e => {
            const due = e.target.value
            if (due) void act(async () => { await api.moveRevision(r.id, due); setMoving(false) })
          }} />
        </label>
      )}
    </li>
  )
}
