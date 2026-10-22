export type Difficulty = 'easy' | 'medium' | 'hard'
export type Outcome = 'solved' | 'stuck'

export interface Attempt {
  id: number
  kind: 'problem' | 'chess'
  url: string
  title: string
  difficulty: Difficulty | ''
  startedAt: string
  endedAt: string | null
  outcome: Outcome | ''
  minutes: number | null
  points: number
}

export interface Award { reason: string; points: number }
export interface Result { awards: Award[]; earned: number; balance: number }

export interface Reward { id: number; name: string; cost: number | null; mysteryWeight: number }
export interface SpendResult { reward: Reward; spent: number; balance: number }

export interface Milestone { id: number; name: string; needed: number; claimedAt: string | null; unlocked: boolean }

export interface Summary {
  balance: number
  todayPoints: number
  week: {
    start: string
    days: { date: string; active: boolean; isToday: boolean; future: boolean }[]
    activeCount: number
    needed: number
    skipPassesLeft: number
    goodWeek: boolean
  }
  goodWeeks: number
  milestones: Milestone[]
  nextReward: { name: string; cost: number; need: number } | null
  running: Attempt | null
}

export interface Redemption { at: string; name: string; spent: number; reason: string }
export interface WeekStat { weekStart: string; solved: number; medianMinutes: number | null }

export class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message)
  }
}

async function call<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch(path, {
    method,
    credentials: 'same-origin',
    headers: {
      'X-DQ': '1',
      ...(body !== undefined ? { 'Content-Type': 'application/json' } : {}),
    },
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })
  let data: unknown = null
  try {
    data = await res.json()
  } catch {
    // empty or non-JSON body
  }
  if (!res.ok) {
    const msg = (data as { error?: string } | null)?.error ?? `Request failed (${res.status})`
    throw new ApiError(res.status, msg)
  }
  return data as T
}

export const api = {
  me: () => call<{ ok: boolean }>('GET', '/api/me'),
  login: (password: string) => call<{ ok: boolean }>('POST', '/api/login', { password }),
  logout: () => call<{ ok: boolean }>('POST', '/api/logout'),
  summary: () => call<Summary>('GET', '/api/summary'),
  start: (url: string, difficulty: Difficulty) => call<Attempt>('POST', '/api/attempts', { url, difficulty }),
  finish: (id: number, outcome: Outcome) => call<Result>('POST', `/api/attempts/${id}/finish`, { outcome }),
  discard: (id: number) => call<{ ok: boolean }>('DELETE', `/api/attempts/${id}`),
  log: (b: { url: string; difficulty: Difficulty; outcome: Outcome; minutes: number; daysAgo: number }) =>
    call<Result>('POST', '/api/attempts/log', b),
  chess: () => call<Result>('POST', '/api/chess'),
  history: () => call<{ attempts: Attempt[]; redemptions: Redemption[] }>('GET', '/api/history'),
  mediumStats: () => call<WeekStat[]>('GET', '/api/stats/mediums'),
  rewards: () => call<Reward[]>('GET', '/api/rewards'),
  createReward: (r: Omit<Reward, 'id'>) => call<{ ok: boolean }>('POST', '/api/rewards', r),
  updateReward: (r: Reward) => call<{ ok: boolean }>('PUT', `/api/rewards/${r.id}`, r),
  removeReward: (id: number) => call<{ ok: boolean }>('DELETE', `/api/rewards/${id}`),
  buy: (id: number) => call<SpendResult>('POST', `/api/rewards/${id}/buy`),
  mystery: () => call<SpendResult>('POST', '/api/mystery'),
  claim: (id: number) => call<{ ok: boolean }>('POST', `/api/milestones/${id}/claim`),
}

export const MYSTERY_COST = 40

const reasonLabels: Record<string, string> = {
  solve: 'Solved',
  fast_bonus: 'Under 25 minutes',
  attempt: 'Honest attempt',
  chess: 'Chess project session',
  good_week: 'Good week bonus',
}

export function reasonLabel(reason: string): string {
  return reasonLabels[reason] ?? reason
}
