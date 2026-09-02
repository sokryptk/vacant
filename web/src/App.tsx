import { useState } from 'react'
import { checkDomains, fetchAgentsMD, fetchTLDs, type CheckResult } from './api'

function Status({ result }: { result: CheckResult }) {
  if (result.error) return <span className="badge">error</span>
  if (result.available) return <span className="badge badge-ok">available</span>
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

function CopyBlock({ text }: { text: string }) {
  const [copied, setCopied] = useState(false)

  async function copy() {
    await navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 1200)
  }

  return (
    <div className="copy-block">
      <pre><code>{text}</code></pre>
      <button className="copy-btn" onClick={copy}>{copied ? 'copied' : 'copy'}</button>
    </div>
  )
}

function skillConfig() {
  const base = window.location.origin
  return JSON.stringify({
    name: "vacant",
    description: "Check domain name availability via DNS and WHOIS",
    endpoint: base + "/api/check",
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: { domains: ["example.com"] },
    response: {
      domain: "string",
      available: "boolean",
      method: "dns | whois",
      reason: "string",
      error: "string"
    }
  }, null, 2)
}

function App() {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<CheckResult[] | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const [apiOpen, setApiOpen] = useState(false)
  const [agentOpen, setAgentOpen] = useState(false)
  const [skillOpen, setSkillOpen] = useState(false)
  const [tldsOpen, setTldsOpen] = useState(false)
  const [agentsMD, setAgentsMD] = useState('')
  const [tlds, setTlds] = useState<string[] | null>(null)
  const [tldsError, setTldsError] = useState('')

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const raw = query.split(/[,\n]+/).map(s => s.trim()).filter(Boolean)
    if (raw.length === 0) return
    setLoading(true)
    setError('')
    setResults(null)
    try {
      const res = await checkDomains(raw.slice(0, 1000))
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

  async function toggleTLDs() {
    if (!tldsOpen && tlds === null) {
      try {
        const data = await fetchTLDs()
        setTlds(data.tlds)
        setTldsError('')
      } catch {
        setTldsError('Unable to load TLDs')
      }
    }
    setTldsOpen(o => !o)
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
          <p className="hint">Separate multiple domains with commas or newlines. Max 1,000.</p>
        </form>

        {error && <div className="error">{error}</div>}

        {results && (
          <div className="results">
            {results.map(r => (
              <div className="result-row" key={r.domain}>
                <span className="result-domain">
                  {!r.available && !r.error ? (
                    <a href={`https://${r.domain}`} target="_blank" rel="noopener noreferrer">
                      {r.domain}
                    </a>
                  ) : (
                    r.domain
                  )}
                </span>
                <span className="result-meta">
                  <Status result={r} />
                  <span className="badge">{r.method}</span>
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
        <pre><code>{`curl -X POST ${window.location.origin}/api/check \\
  -H "Content-Type: application/json" \\
  -d '{"domain":"example.com"}'`}</code></pre>
        <p>Batch check:</p>
        <pre><code>{`curl -X POST ${window.location.origin}/api/check \\
  -H "Content-Type: application/json" \\
  -d '{"domains":["foo.com","bar.io","baz.dev"]}'`}</code></pre>
      </Collapsible>

      <Collapsible title="Add to your agent" open={skillOpen} onToggle={() => setSkillOpen(o => !o)}>
        <p>Copy this block into your agent's skill/tool config:</p>
        <CopyBlock text={skillConfig()} />
        <p style={{ marginTop: '0.75rem', fontSize: '0.85rem', color: '#555' }}>
          Point <code>endpoint</code> to wherever you run vacant.
        </p>
      </Collapsible>

      <Collapsible title="Agent Access" open={agentOpen} onToggle={toggleAgent}>
        <p style={{ marginBottom: '0.75rem', fontSize: '0.85rem', color: '#555' }}>
          Copy and paste this into your agent's chat to install the skill:
        </p>
        <CopyBlock text={agentsMD} />
      </Collapsible>

      <Collapsible title="Available TLDs" open={tldsOpen} onToggle={toggleTLDs}>
        {tldsError && <div className="error">{tldsError}</div>}
        {tlds && (
          <div>
            <p style={{ marginTop: 0, fontSize: '0.85rem', color: '#555' }}>
              {tlds.length.toLocaleString()} TLDs from IANA
            </p>
            <pre className="wrap" style={{ maxHeight: '300px', overflowY: 'auto' }}>
              {tlds.join(', ')}
            </pre>
          </div>
        )}
      </Collapsible>
    </div>
  )
}

export default App
