export interface CheckResult {
  domain: string
  available: boolean
  method: string
  reason?: string
  error?: string
}

async function postJSON(path: string, body: unknown): Promise<unknown> {
  const res = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || `HTTP ${res.status}`)
  }
  return res.json()
}

export async function checkDomains(domains: string[]): Promise<CheckResult[]> {
  const data = await postJSON('/api/check', { domains })
  return (data as { results: CheckResult[] }).results
}

export async function fetchAgentsMD(): Promise<string> {
  const res = await fetch('/agents.md')
  if (!res.ok) throw new Error('Failed to load agents.md')
  return res.text()
}

export async function fetchTLDs(): Promise<{ count: number; tlds: string[] }> {
  const res = await fetch('/api/tlds')
  if (!res.ok) throw new Error('Failed to load TLDs')
  return res.json() as Promise<{ count: number; tlds: string[] }>
}
