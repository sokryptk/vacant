import { useState } from 'react'
import type { FormEvent } from 'react'
import { checkDomains, fetchAgentsMD, type CheckResult } from './api'

function StatusBadge({ result }: { result: CheckResult }) {
  if (result.error) {
    return <span className="badge badge-error" title={result.error}>error</span>
  }
  if (result.available) {
    return <span className="badge badge-ok">available</span>
  }
  return <span className="badge badge-taken">taken</span>
}

function Collapsible({ title, open, onToggle, children }: {
  title: string
  open: boolean
  onToggle: () => void
  children: React.ReactNode
}) {
  return (
    <div className="card">
      <div className="collapsible-header" onClick={onToggle}>
        <span className="section-title">{title}</span>
        <span className={`chevron ${open ? 'open' : ''}`}>▶</span>
      </div>
      {open && children}
    </div>
  )
}

function App() {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<CheckResult[] | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const [apiOpen, setApiOpen] = useState(false)
  const [agentOpen, setAgentOpen] = useState(false)
  const [agentsMD, setAgentsMD] = useState('')

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    const raw = query.split(/[,\n]+/).map(s => s.trim()).filter(Boolean)
    if (raw.length === 0) return
    setLoading(true)
    setError('')
    setResults(null)
    try {
      const res = await checkDomains(raw.slice(0, 50))
      setResults(res)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Check failed')
    } finally {
      setLoading(false)
    }
  }

  async function toggleAgent() {
    if (!agentOpen && !agentsMD) {
      try {
        const md = await fetchAgentsMD()
        setAgentsMD(md)
      } catch {
        setAgentsMD('Unable to load agents.md')
      }
    }
    setAgentOpen(o => !o)
  }

  return (
    <div className="container">
      <h1>vacant</h1>
      <p className="subtitle">Check if a domain name is available for registration.</p>

      <div className="card">
        <form onSubmit={handleSubmit}>
          <div className="input-group">
            <input
              type="text"
              placeholder="example.com, example.io ..."
              value={query}
              onChange={e => setQuery(e.target.value)}
              disabled={loading}
            />
            <button type="submit" disabled={loading}>
              {loading ? 'Checking…' : 'Check'}
            </button>
          </div>
          <p className="hint">Separate multiple domains with commas or newlines. Max 50.</p>
        </form>

        {error && <div className="error">{error}</div>}

        {results && (
          <div className="results">
            {results.map(r => (
              <div className="result-row" key={r.domain}>
                <span className="result-domain">{r.domain}</span>
                <span className="result-meta">
                  <StatusBadge result={r} />
                  <span className="badge badge-method">{r.method}</span>
                  {r.reason && <span className="reason">{r.reason}</span>}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>

      <Collapsible title="API Access" open={apiOpen} onToggle={() => setApiOpen(o => !o)}>
        <p>POST <code>/api/check</code> with a JSON body.</p>
        <p>Single domain:</p>
        <pre><code>{`curl -X POST http://localhost:8080/api/check \\
  -H "Content-Type: application/json" \\
  -d '{"domain":"example.com"}'`}</code></pre>
        <p>Batch check:</p>
        <pre><code>{`curl -X POST http://localhost:8080/api/check \\
  -H "Content-Type: application/json" \\
  -d '{"domains":["foo.com","bar.io","baz.dev"]}'`}</code></pre>
      </Collapsible>

      <Collapsible title="Agent Access" open={agentOpen} onToggle={toggleAgent}>
        <pre className="wrap">{agentsMD}</pre>
      </Collapsible>
    </div>
  )
}

export default App
