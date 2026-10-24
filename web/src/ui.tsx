import { reasonLabel, type Difficulty, type Result } from './api'

export type Toast = { title: string; lines?: string[] } | null

export const DIFFICULTIES: Difficulty[] = ['easy', 'medium', 'hard']

export function cap(s: string) {
  return s.charAt(0).toUpperCase() + s.slice(1)
}

export function errText(e: unknown) {
  return e instanceof Error ? e.message : 'Something went wrong'
}

export function resultToast(r: Result): Toast {
  return {
    title: `+${r.earned} points`,
    lines: r.awards.map(a => `${reasonLabel(a.reason)}  +${a.points}`),
  }
}

export function DifficultyPicker({ value, onChange }: { value: Difficulty; onChange: (d: Difficulty) => void }) {
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

// ---------- calendar dates ----------
// Dates are game days as YYYY-MM-DD strings. All arithmetic is done in UTC so
// the viewer's time zone can never shift a day.

export function parseDay(s: string) {
  return new Date(s + 'T00:00:00Z')
}

export function isoDay(d: Date) {
  return d.toISOString().slice(0, 10)
}

export function addDays(s: string, n: number) {
  const d = parseDay(s)
  d.setUTCDate(d.getUTCDate() + n)
  return isoDay(d)
}

export function localToday() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

export function fmtDay(s: string) {
  return parseDay(s).toLocaleDateString(undefined, { weekday: 'short', day: 'numeric', month: 'short', timeZone: 'UTC' })
}
