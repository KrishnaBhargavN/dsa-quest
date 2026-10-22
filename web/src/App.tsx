import { useCallback, useEffect, useRef, useState } from 'react'
import {
  api, ApiError, MYSTERY_COST, reasonLabel,
  type Attempt, type Difficulty, type Milestone, type Outcome, type Redemption,
  type Result, type Reward, type Summary, type WeekStat,
} from './api'

type Tab = 'home' | 'solve' | 'shop' | 'history'
type Toast = { title: string; lines?: string[] } | null

const DAY_LETTERS = ['M', 'T', 'W', 'T', 'F', 'S', 'S']
const DIFFICULTIES: Difficulty[] = ['easy', 'medium', 'hard']

function cap(s: string) {
  return s.charAt(0).toUpperCase() + s.slice(1)
}

function errText(e: unknown) {
  return e instanceof Error ? e.message : 'Something went wrong'
}

function resultToast(r: Result): Toast {
  return {
    title: `+${r.earned} points`,
    lines: r.awards.map(a => `${reasonLabel(a.reason)}  +${a.points}`),
  }
}

export default function App() {
  const [auth, setAuth] = useState<'checking' | 'out' | 'in'>('checking')
  const [fatal, setFatal] = useState('')
  const [tab, setTab] = useState<Tab>('home')
  const [summary, setSummary] = useState<Summary | null>(null)
  const [toast, setToast] = useState<Toast>(null)
  const toastTimer = useRef<number | undefined>(undefined)

  const showToast = useCallback((t: Toast) => {
    window.clearTimeout(toastTimer.current)
    setToast(t)
    toastTimer.current = window.setTimeout(() => setToast(null), 4200)
  }, [])

  const refresh = useCallback(async () => {
    try {
      setSummary(await api.summary())
    } catch (e) {
      if (e instanceof ApiError && e.status === 401) setAuth('out')
      else showToast({ title: errText(e) })
    }
  }, [showToast])

  useEffect(() => {
    api.me()
      .then(() => setAuth('in'))
      .catch(e => {
        if (e instanceof ApiError && e.status === 401) setAuth('out')
        else setFatal(errText(e))
      })
  }, [])

  useEffect(() => {
    if (auth === 'in') void refresh()
  }, [auth, refresh])

  useEffect(() => {
    const onFocus = () => { if (auth === 'in') void refresh() }
    window.addEventListener('focus', onFocus)
    return () => window.removeEventListener('focus', onFocus)
  }, [auth, refresh])

  if (fatal) {
    return (
      <main className="center-screen">
        <h1 className="wordmark">DSA Quest</h1>
        <p className="muted">{fatal}</p>
        <button className="btn" onClick={() => location.reload()}>Try again</button>
      </main>
    )
  }
  if (auth === 'checking') return <main className="center-screen"><p className="muted">Loading</p></main>
  if (auth === 'out') return <Login onIn={() => setAuth('in')} />

  return (
    <div className="app">
      <header className="topbar">
        <span className="wordmark">DSA Quest</span>
        <button className="link" onClick={async () => { await api.logout(); setAuth('out') }}>Log out</button>
      </header>
      <main className="screen">
        {!summary ? <p className="muted">Loading</p> : (
          <>
            {tab === 'home' && <Home s={summary} go={setTab} refresh={refresh} toast={showToast} />}
            {tab === 'solve' && <Solve s={summary} refresh={refresh} toast={showToast} />}
            {tab === 'shop' && <Shop s={summary} refresh={refresh} toast={showToast} />}
            {tab === 'history' && <History />}
          </>
        )}
      </main>
      {toast && (
        <div className="toast" role="status" onClick={() => setToast(null)}>
          <strong>{toast.title}</strong>
          {toast.lines?.map(l => <span key={l}>{l}</span>)}
        </div>
      )}
      <nav className="tabbar" aria-label="Sections">
        {(['home', 'solve', 'shop', 'history'] as Tab[]).map(t => (
          <button key={t} className={tab === t ? 'tab on' : 'tab'} onClick={() => setTab(t)} aria-current={tab === t ? 'page' : undefined}>
            {t === 'solve' && summary?.running ? <span className="dot" aria-label="timer running" /> : null}
            {cap(t)}
          </button>
        ))}
      </nav>
    </div>
  )
}

// ---------------------------------------------------------------- login

function Login({ onIn }: { onIn: () => void }) {
  const [pw, setPw] = useState('')
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)
  return (
    <main className="center-screen">
      <h1 className="wordmark big">DSA Quest</h1>
      <form
        className="stack"
        onSubmit={async e => {
          e.preventDefault()
          setBusy(true)
          setErr('')
          try {
            await api.login(pw)
            onIn()
          } catch (ex) {
            setErr(errText(ex))
          } finally {
            setBusy(false)
          }
        }}
      >
        <label className="field">
          <span>Password</span>
          <input type="password" autoComplete="current-password" value={pw} onChange={e => setPw(e.target.value)} autoFocus />
        </label>
        {err && <p className="error">{err}</p>}
        <button className="btn primary" disabled={busy || !pw}>{busy ? 'Checking' : 'Log in'}</button>
      </form>
    </main>
  )
}

// ---------------------------------------------------------------- home

type ScreenProps = { s: Summary; refresh: () => Promise<void>; toast: (t: Toast) => void }

function Home({ s, go, refresh, toast }: ScreenProps & { go: (t: Tab) => void }) {
  const [busy, setBusy] = useState(false)
  const next = s.nextReward
  const progress = next ? Math.max(0, Math.min(1, s.balance / next.cost)) : 1
  const target = s.milestones.find(m => !m.claimedAt)

  return (
    <div className="stack-lg">
      <section className="balance" aria-label="Points balance">
        <div className="balance-num">{s.balance}</div>
        <div className="balance-unit">points</div>
        <p className="today-line">{s.todayPoints > 0 ? `+${s.todayPoints} earned today` : 'Nothing logged today yet'}</p>
        <div className="meter" role="progressbar" aria-valuemin={0} aria-valuemax={100} aria-valuenow={Math.round(progress * 100)}>
          <div className="meter-fill" style={{ width: `${progress * 100}%` }} />
        </div>
        <p className="muted small">
          {next ? `Next: ${next.name}, ${next.need} points away` : 'You can afford everything in the shop'}
        </p>
      </section>

      {s.running && (
        <button className="running" onClick={() => go('solve')}>
          <span>Timer running</span>
          <strong>{s.running.title || cap(s.running.difficulty)}</strong>
          <Elapsed since={s.running.startedAt} />
        </button>
      )}

      <WeekBoard s={s} />

      {target && <QueenTrack m={target} have={s.goodWeeks} />}

      <div className="actions">
        <button className="btn primary" onClick={() => go('solve')}>
          {s.running ? 'Back to the timer' : 'Start a problem'}
        </button>
        <button
          className="btn"
          disabled={busy}
          onClick={async () => {
            setBusy(true)
            try {
              toast(resultToast(await api.chess()))
              await refresh()
            } catch (e) {
              toast({ title: errText(e) })
            } finally {
              setBusy(false)
            }
          }}
        >
          Log chess session
        </button>
      </div>
    </div>
  )
}

function WeekBoard({ s }: { s: Summary }) {
  const w = s.week
  const piece = w.goodWeek ? '♛' : '♟'
  let caption: string
  if (w.goodWeek) caption = `Good week banked, +50. ${w.activeCount} active days so far.`
  else {
    const left = w.needed - w.activeCount
    caption = `${w.activeCount} of ${w.needed} days for a good week. ${w.skipPassesLeft} skip ${w.skipPassesLeft === 1 ? 'pass' : 'passes'} left.`
    if (left > 0 && w.skipPassesLeft === 0 && w.days.filter(d => d.future || d.isToday).length < left) {
      caption = `${w.activeCount} of ${w.needed} days. This week can't make it now, so bank what you can.`
    }
  }
  return (
    <section aria-label="This week">
      <div className="board">
        {w.days.map((d, i) => {
          const missed = !d.active && !d.future && !d.isToday
          const cls = ['sq', i % 2 === 0 ? 'light' : 'dark', d.isToday ? 'today' : '', d.future ? 'future' : ''].join(' ')
          return (
            <div key={d.date} className={cls} aria-label={`${d.date}${d.active ? ', active' : missed ? ', skipped' : ''}`}>
              <span className="piece">{d.active ? piece : missed ? '·' : ''}</span>
              <span className="file">{DAY_LETTERS[i]}</span>
            </div>
          )
        })}
      </div>
      <p className="muted small">{caption}</p>
    </section>
  )
}

function QueenTrack({ m, have }: { m: Milestone; have: number }) {
  const owned = Math.min(have, m.needed)
  return (
    <section className="queens" aria-label="Next big reward">
      <div className="queen-row" aria-hidden="true">
        {Array.from({ length: m.needed }, (_, i) => (
          <span key={i} className={i < owned ? 'q owned' : 'q'}>{i < owned ? '♛' : '♕'}</span>
        ))}
      </div>
      <p className="small">
        {m.unlocked
          ? `${m.name} unlocked. Claim it in the shop.`
          : `${m.name}: ${owned} of ${m.needed} good weeks`}
      </p>
    </section>
  )
}

function Elapsed({ since, big }: { since: string; big?: boolean }) {
  const [now, setNow] = useState(() => Date.now())
  useEffect(() => {
    const t = window.setInterval(() => setNow(Date.now()), 1000)
    return () => window.clearInterval(t)
  }, [])
  const secs = Math.max(0, Math.floor((now - new Date(since).getTime()) / 1000))
  const h = Math.floor(secs / 3600)
  const m = Math.floor((secs % 3600) / 60)
  const sec = secs % 60
  const text = h > 0
    ? `${h}:${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}`
    : `${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}`
  return <span className={big ? 'clock big' : 'clock'} aria-label={`${m} minutes elapsed`}>{text}</span>
}

// ---------------------------------------------------------------- solve

function DifficultyPicker({ value, onChange }: { value: Difficulty; onChange: (d: Difficulty) => void }) {
  return (
    <div className="seg" role="radiogroup" aria-label="Difficulty">
      {DIFFICULTIES.map(d => (
        <button key={d} type="button" role="radio" aria-checked={value === d} className={value === d ? `on ${d}` : d} onClick={() => onChange(d)}>
          {cap(d)}
        </button>
      ))}
    </div>
  )
}

function Solve({ s, refresh, toast }: ScreenProps) {
  const [url, setUrl] = useState('')
  const [difficulty, setDifficulty] = useState<Difficulty>('medium')
  const [busy, setBusy] = useState(false)
  const [manual, setManual] = useState(false)
  const [minutes, setMinutes] = useState('')
  const [outcome, setOutcome] = useState<Outcome>('solved')
  const [daysAgo, setDaysAgo] = useState(0)
  const run = s.running

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

  if (run) {
    return (
      <div className="stack-lg">
        <section className="timer">
          <p className="muted small">{cap(run.difficulty)}</p>
          <h2 className="problem-title">
            {run.url ? <a href={run.url} target="_blank" rel="noreferrer">{run.title || run.url}</a> : (run.title || 'Untitled problem')}
          </h2>
          <Elapsed since={run.startedAt} big />
          {run.difficulty === 'medium' && <p className="muted small">Solve a medium in under 25:00 for +10.</p>}
        </section>
        <div className="actions">
          <button className="btn primary" disabled={busy} onClick={() => act(async () => {
            toast(resultToast(await api.finish(run.id, 'solved')))
            await refresh()
          })}>Solved</button>
          <button className="btn" disabled={busy} onClick={() => act(async () => {
            toast(resultToast(await api.finish(run.id, 'stuck')))
            await refresh()
          })}>Stuck, log the attempt</button>
          <button className="link danger" disabled={busy} onClick={() => act(async () => {
            if (!confirm('Discard this timer? Nothing will be logged.')) return
            await api.discard(run.id)
            toast({ title: 'Timer discarded' })
            await refresh()
          })}>Discard timer</button>
        </div>
      </div>
    )
  }

  return (
    <div className="stack-lg">
      <form
        className="stack"
        onSubmit={e => {
          e.preventDefault()
          void act(async () => {
            await api.start(url, difficulty)
            setUrl('')
            await refresh()
          })
        }}
      >
        <h2 className="section-title">Start a problem</h2>
        <label className="field">
          <span>Problem link (optional)</span>
          <input type="url" inputMode="url" placeholder="https://leetcode.com/problems/…" value={url} onChange={e => setUrl(e.target.value)} />
        </label>
        <DifficultyPicker value={difficulty} onChange={setDifficulty} />
        <button className="btn primary" disabled={busy}>Start timer</button>
      </form>

      <section className="stack">
        <button className="link" aria-expanded={manual} onClick={() => setManual(m => !m)}>
          {manual ? 'Hide' : 'Solved it without the timer? Log it here'}
        </button>
        {manual && (
          <form
            className="stack panel"
            onSubmit={e => {
              e.preventDefault()
              const mins = parseInt(minutes, 10)
              if (!mins || mins < 1) {
                toast({ title: 'Enter how many minutes it took' })
                return
              }
              void act(async () => {
                toast(resultToast(await api.log({ url, difficulty, outcome, minutes: mins, daysAgo })))
                setMinutes('')
                setUrl('')
                await refresh()
              })
            }}
          >
            <p className="muted small">Uses the link and difficulty above.</p>
            <label className="field">
              <span>Minutes it took</span>
              <input type="number" inputMode="numeric" min={1} max={600} value={minutes} onChange={e => setMinutes(e.target.value)} />
            </label>
            <div className="seg" role="radiogroup" aria-label="Outcome">
              {(['solved', 'stuck'] as Outcome[]).map(o => (
                <button key={o} type="button" role="radio" aria-checked={outcome === o} className={outcome === o ? 'on' : ''} onClick={() => setOutcome(o)}>
                  {cap(o)}
                </button>
              ))}
            </div>
            <div className="seg" role="radiogroup" aria-label="When">
              {[0, 1].map(d => (
                <button key={d} type="button" role="radio" aria-checked={daysAgo === d} className={daysAgo === d ? 'on' : ''} onClick={() => setDaysAgo(d)}>
                  {d === 0 ? 'Today' : 'Yesterday'}
                </button>
              ))}
            </div>
            <button className="btn" disabled={busy}>Log it</button>
          </form>
        )}
      </section>
    </div>
  )
}

// ---------------------------------------------------------------- shop

function Shop({ s, refresh, toast }: ScreenProps) {
  const [rewards, setRewards] = useState<Reward[] | null>(null)
  const [armed, setArmed] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)
  const [editing, setEditing] = useState(false)
  const [reveal, setReveal] = useState<Reward | null>(null)

  const load = useCallback(async () => {
    try {
      setRewards(await api.rewards())
    } catch (e) {
      toast({ title: errText(e) })
    }
  }, [toast])

  useEffect(() => { void load() }, [load])

  useEffect(() => {
    if (!armed) return
    const t = window.setTimeout(() => setArmed(null), 3000)
    return () => window.clearTimeout(t)
  }, [armed])

  // First tap arms the button, second tap spends. Stops accidental purchases.
  const spend = async (key: string, fn: () => Promise<void>) => {
    if (armed !== key) {
      setArmed(key)
      return
    }
    setArmed(null)
    setBusy(true)
    try {
      await fn()
      await refresh()
    } catch (e) {
      toast({ title: errText(e) })
    } finally {
      setBusy(false)
    }
  }

  if (!rewards) return <p className="muted">Loading the shop</p>
  if (editing) return <ShopEditor rewards={rewards} done={async () => { setEditing(false); await load() }} toast={toast} />

  const buyable = rewards.filter(r => r.cost !== null)
  const pool = rewards.filter(r => r.mysteryWeight > 0)
  const poolTotal = pool.reduce((n, r) => n + r.mysteryWeight, 0)

  return (
    <div className="stack-lg">
      <div className="shop-head">
        <h2 className="section-title">Shop</h2>
        <span className="pill">{s.balance} points</span>
      </div>

      <ul className="rows">
        {buyable.map(r => {
          const key = `buy-${r.id}`
          const afford = s.balance >= (r.cost ?? 0)
          return (
            <li key={r.id} className="row">
              <span className="row-name">{r.name}</span>
              <span className="cost">{r.cost}</span>
              <button className={armed === key ? 'btn small armed' : 'btn small'} disabled={!afford || busy}
                onClick={() => spend(key, async () => {
                  const res = await api.buy(r.id)
                  toast({ title: `Bought: ${res.reward.name}`, lines: [`${res.balance} points left. Enjoy it.`] })
                })}>
                {armed === key ? 'Tap to confirm' : 'Buy'}
              </button>
            </li>
          )
        })}
        <li className="row mystery">
          <span className="row-name">
            Mystery box
            <span className="muted small block">
              {pool.map(r => `${r.name} ${Math.round((r.mysteryWeight / poolTotal) * 100)}%`).join(', ')}
            </span>
          </span>
          <span className="cost">{MYSTERY_COST}</span>
          <button className={armed === 'mystery' ? 'btn small armed' : 'btn small'} disabled={s.balance < MYSTERY_COST || busy || pool.length === 0}
            onClick={() => spend('mystery', async () => {
              const res = await api.mystery()
              setReveal(res.reward)
            })}>
            {armed === 'mystery' ? 'Tap to confirm' : 'Open'}
          </button>
        </li>
      </ul>

      <section className="stack">
        <h2 className="section-title">Big rewards</h2>
        <p className="muted small">Unlocked by good weeks, not points. {s.goodWeeks} good {s.goodWeeks === 1 ? 'week' : 'weeks'} banked so far.</p>
        <ul className="rows">
          {s.milestones.map(m => (
            <li key={m.id} className="row">
              <span className="row-name">
                {m.name}
                <span className="muted small block">{m.needed} good weeks</span>
              </span>
              {m.claimedAt ? (
                <span className="muted small">Claimed {new Date(m.claimedAt).toLocaleDateString(undefined, { day: 'numeric', month: 'short' })}</span>
              ) : m.unlocked ? (
                <button className={armed === `m-${m.id}` ? 'btn small armed' : 'btn small primary'} disabled={busy}
                  onClick={() => spend(`m-${m.id}`, async () => {
                    await api.claim(m.id)
                    toast({ title: `${m.name} claimed`, lines: ['Go get it. You earned it.'] })
                  })}>
                  {armed === `m-${m.id}` ? 'Tap to confirm' : 'Claim'}
                </button>
              ) : (
                <span className="muted small">{Math.min(s.goodWeeks, m.needed)} of {m.needed}</span>
              )}
            </li>
          ))}
        </ul>
      </section>

      <button className="link" onClick={() => setEditing(true)}>Edit rewards and prices</button>

      {reveal && (
        <div className="overlay" role="dialog" aria-label="Mystery box result" onClick={() => setReveal(null)}>
          <div className="reveal">
            <p className="muted small">The box holds</p>
            <p className="reveal-name">{reveal.name}</p>
            <button className="btn primary" onClick={() => setReveal(null)}>Nice</button>
          </div>
        </div>
      )}
    </div>
  )
}

type Draft = { id: number | null; name: string; cost: string; weight: string }

function ShopEditor({ rewards, done, toast }: { rewards: Reward[]; done: () => Promise<void>; toast: (t: Toast) => void }) {
  const [rows, setRows] = useState<Draft[]>(() => rewards.map(r => ({
    id: r.id, name: r.name, cost: r.cost === null ? '' : String(r.cost), weight: String(r.mysteryWeight),
  })))
  const [busy, setBusy] = useState(false)

  const set = (i: number, patch: Partial<Draft>) => setRows(rs => rs.map((r, j) => (j === i ? { ...r, ...patch } : r)))

  const save = async () => {
    setBusy(true)
    try {
      const original = new Map(rewards.map(r => [r.id, r]))
      for (const d of rows) {
        if (!d.name.trim()) continue
        const cost = d.cost.trim() === '' ? null : parseInt(d.cost, 10)
        const weight = parseInt(d.weight || '0', 10) || 0
        if (cost !== null && (!Number.isFinite(cost) || cost < 1)) throw new Error(`Price for "${d.name}" must be a whole number above 0, or empty`)
        if (cost === null && weight === 0) throw new Error(`"${d.name}" needs a price or a mystery weight`)
        const payload = { name: d.name.trim(), cost, mysteryWeight: weight }
        if (d.id === null) {
          await api.createReward(payload)
        } else {
          const o = original.get(d.id)
          if (!o || o.name !== payload.name || o.cost !== payload.cost || o.mysteryWeight !== payload.mysteryWeight) {
            await api.updateReward({ id: d.id, ...payload })
          }
        }
      }
      toast({ title: 'Shop saved' })
      await done()
    } catch (e) {
      toast({ title: errText(e) })
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="stack-lg">
      <h2 className="section-title">Edit rewards</h2>
      <p className="muted small">Leave the price empty for a mystery-box-only prize. Weight sets how often the box draws it; 0 keeps it out of the box.</p>
      <div className="editor">
        <div className="editor-head" aria-hidden="true"><span>Reward</span><span>Price</span><span>Weight</span><span /></div>
        {rows.map((d, i) => (
          <div key={d.id ?? `new-${i}`} className="editor-row">
            <input aria-label="Reward name" value={d.name} onChange={e => set(i, { name: e.target.value })} />
            <input aria-label="Price" inputMode="numeric" value={d.cost} onChange={e => set(i, { cost: e.target.value.replace(/\D/g, '') })} />
            <input aria-label="Mystery weight" inputMode="numeric" value={d.weight} onChange={e => set(i, { weight: e.target.value.replace(/\D/g, '') })} />
            <button className="link danger" aria-label={`Remove ${d.name}`} disabled={busy} onClick={async () => {
              if (d.id === null) { setRows(rs => rs.filter((_, j) => j !== i)); return }
              if (!confirm(`Remove "${d.name}" from the shop?`)) return
              try {
                await api.removeReward(d.id)
                setRows(rs => rs.filter((_, j) => j !== i))
              } catch (e) {
                toast({ title: errText(e) })
              }
            }}>Remove</button>
          </div>
        ))}
      </div>
      <button className="link" onClick={() => setRows(rs => [...rs, { id: null, name: '', cost: '', weight: '0' }])}>Add a reward</button>
      <div className="actions">
        <button className="btn primary" disabled={busy} onClick={save}>Save shop</button>
        <button className="btn" disabled={busy} onClick={() => void done()}>Cancel</button>
      </div>
    </div>
  )
}

// ---------------------------------------------------------------- history

function History() {
  const [data, setData] = useState<{ attempts: Attempt[]; redemptions: Redemption[] } | null>(null)
  const [stats, setStats] = useState<WeekStat[] | null>(null)
  const [err, setErr] = useState('')

  useEffect(() => {
    Promise.all([api.history(), api.mediumStats()])
      .then(([h, st]) => { setData(h); setStats(st) })
      .catch(e => setErr(errText(e)))
  }, [])

  if (err) return <p className="error">{err}</p>
  if (!data || !stats) return <p className="muted">Loading history</p>

  const fmtDate = (iso: string) => new Date(iso).toLocaleString(undefined, { day: 'numeric', month: 'short', hour: 'numeric', minute: '2-digit' })

  return (
    <div className="stack-lg">
      <MediumChart stats={stats} />

      <section className="stack">
        <h2 className="section-title">Recent problems</h2>
        {data.attempts.length === 0 ? (
          <p className="muted small">Nothing yet. Start a problem and it shows up here.</p>
        ) : (
          <ul className="rows">
            {data.attempts.map(a => (
              <li key={a.id} className="row">
                <span className="row-name">
                  {a.kind === 'chess' ? 'Chess project session' : (a.title || cap(a.difficulty))}
                  <span className="muted small block">
                    {a.kind === 'chess'
                      ? fmtDate(a.startedAt)
                      : `${cap(a.difficulty)}, ${a.outcome === 'solved' ? 'solved' : 'stuck'} in ${Math.round(a.minutes ?? 0)} min. ${fmtDate(a.startedAt)}`}
                  </span>
                </span>
                <span className="cost earned">+{a.points}</span>
              </li>
            ))}
          </ul>
        )}
      </section>

      {data.redemptions.length > 0 && (
        <section className="stack">
          <h2 className="section-title">Rewards taken</h2>
          <ul className="rows">
            {data.redemptions.map((r, i) => (
              <li key={i} className="row">
                <span className="row-name">
                  {r.name}
                  <span className="muted small block">{r.reason === 'mystery' ? 'From the mystery box. ' : ''}{fmtDate(r.at)}</span>
                </span>
                <span className="cost">-{r.spent}</span>
              </li>
            ))}
          </ul>
        </section>
      )}
    </div>
  )
}

function MediumChart({ stats }: { stats: WeekStat[] }) {
  const W = 340, H = 170, top = 18, bottom = 34, left = 8
  const maxVal = Math.max(40, ...stats.map(s => s.medianMinutes ?? 0))
  const plotH = H - top - bottom
  const bw = (W - left * 2) / stats.length
  const y = (v: number) => top + plotH - (v / maxVal) * plotH
  const any = stats.some(s => s.medianMinutes !== null)
  return (
    <section className="stack">
      <h2 className="section-title">Medium solve time</h2>
      <p className="muted small">Median minutes per week for solved mediums. The line marks 25 minutes, the bonus cutoff.</p>
      {!any ? (
        <p className="muted small">Solve a medium with the timer to start this chart.</p>
      ) : (
        <svg className="chart" viewBox={`0 0 ${W} ${H}`} role="img" aria-label="Median minutes for solved mediums, last 8 weeks">
          <line x1={left} x2={W - left} y1={y(25)} y2={y(25)} className="ref" />
          {stats.map((s, i) => {
            const x = left + i * bw + bw * 0.18
            const w = bw * 0.64
            const d = new Date(s.weekStart + 'T00:00:00')
            const label = d.toLocaleDateString(undefined, { day: 'numeric', month: 'short' })
            return (
              <g key={s.weekStart}>
                {s.medianMinutes !== null && (
                  <>
                    <rect x={x} y={y(s.medianMinutes)} width={w} height={top + plotH - y(s.medianMinutes)} rx="3"
                      className={s.medianMinutes < 25 ? 'bar fast' : 'bar'} />
                    <text x={x + w / 2} y={y(s.medianMinutes) - 5} textAnchor="middle" className="val">{Math.round(s.medianMinutes)}</text>
                  </>
                )}
                <text x={x + w / 2} y={H - 12} textAnchor="middle" className="axis">{label}</text>
              </g>
            )
          })}
        </svg>
      )}
    </section>
  )
}
