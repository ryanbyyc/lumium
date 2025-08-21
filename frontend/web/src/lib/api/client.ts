export type ApiError = {
  status: number
  code?: string
  message: string
  details?: unknown
}

type FetcherOpts = {
  method?: "GET" | "POST" | "PUT" | "PATCH" | "DELETE"
  headers?: Record<string, string>
  body?: any
  signal?: AbortSignal
  credentials?: RequestCredentials
}

const BASE = process.env.NEXT_PUBLIC_API_URL ?? "https://api.tst.com"

async function parse<T>(res: Response): Promise<T> {
  const txt = await res.text()
  if (!txt) return undefined as unknown as T
  try {
    return JSON.parse(txt) as T
  } catch {
    return txt as unknown as T
  }
}

async function doFetch<T>(path: string, opts: FetcherOpts = {}): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method: opts.method ?? "GET",
    headers: { "content-type": "application/json", ...(opts.headers ?? {}) },
    credentials: opts.credentials ?? "include",
    body: opts.body ? JSON.stringify(opts.body) : undefined,
    signal: opts.signal,
  })

  if (res.ok) return parse<T>(res)

  const payload = await parse<ApiError>(res).catch(() => undefined)
  const err: ApiError = {
    status: res.status,
    code: payload?.code,
    message: payload?.message ?? res.statusText,
    details: payload?.details,
  }
  throw err
}

// Auto-refresh wrapper: retries once on 401 using /auth/refresh
let refreshing = Promise.resolve()
async function withRefresh<T>(fn: () => Promise<T>): Promise<T> {
  try {
    return await fn()
  } catch (e: any) {
    if (e?.status !== 401) throw e
    // gate concurrent refreshes
    if (!refreshing) {
      refreshing = doFetch("/auth/refresh", { method: "POST" })
        .catch(() => Promise.reject(e))
        .finally(() => {
          refreshing = Promise.resolve()
        })
    }
    await refreshing
    return fn()
  }
}

export const api = {
  get: <T>(path: string, opts?: FetcherOpts) => withRefresh(() => doFetch<T>(path, opts)),
  post: <T>(path: string, body?: any, opts?: FetcherOpts) =>
    withRefresh(() => doFetch<T>(path, { ...opts, method: "POST", body })),
  put: <T>(path: string, body?: any, opts?: FetcherOpts) =>
    withRefresh(() => doFetch<T>(path, { ...opts, method: "PUT", body })),
  patch: <T>(path: string, body?: any, opts?: FetcherOpts) =>
    withRefresh(() => doFetch<T>(path, { ...opts, method: "PATCH", body })),
  del: <T>(path: string, opts?: FetcherOpts) =>
    withRefresh(() => doFetch<T>(path, { ...opts, method: "DELETE" })),
}
