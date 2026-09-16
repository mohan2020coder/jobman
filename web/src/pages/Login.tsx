import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { errorMessage } from '../api/client'

export function LoginPage() {
  const { login, register, user } = useAuth()
  const navigate = useNavigate()
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const [phone, setPhone] = useState('9876543210')
  const [password, setPassword] = useState('password123')
  const [businessName, setBusinessName] = useState('')
  const [ownerName, setOwnerName] = useState('')
  const [email, setEmail] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  if (user) {
    navigate('/dashboard', { replace: true })
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setBusy(true)
    try {
      if (mode === 'login') {
        await login(phone, password)
      } else {
        await register({
          business_name: businessName,
          owner_name: ownerName,
          phone,
          email,
          password,
        })
      }
      navigate('/dashboard', { replace: true })
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="auth-wrap">
      <div className="auth-card">
        <h1 className="auth-title">Service Manager</h1>
        <p className="auth-sub">Job management for field-service businesses</p>

        <div className="auth-tabs">
          <button className={mode === 'login' ? 'tab active' : 'tab'} onClick={() => setMode('login')}>
            Sign in
          </button>
          <button className={mode === 'register' ? 'tab active' : 'tab'} onClick={() => setMode('register')}>
            Register business
          </button>
        </div>

        <form onSubmit={handleSubmit} className="auth-form">
          {mode === 'register' && (
            <>
              <label className="field">
                <span className="field-label">Business name</span>
                <input
                  required
                  value={businessName}
                  onChange={(e) => setBusinessName(e.target.value)}
                  placeholder="ABC AC Services"
                />
              </label>
              <label className="field">
                <span className="field-label">Your name</span>
                <input
                  required
                  value={ownerName}
                  onChange={(e) => setOwnerName(e.target.value)}
                  placeholder="Rajesh"
                />
              </label>
              <label className="field">
                <span className="field-label">Email (optional)</span>
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="owner@example.com"
                />
              </label>
            </>
          )}
          <label className="field">
            <span className="field-label">Phone</span>
            <input required value={phone} onChange={(e) => setPhone(e.target.value)} placeholder="9876543210" />
          </label>
          <label className="field">
            <span className="field-label">Password</span>
            <input
              required
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
            />
          </label>

          {error && <div className="alert alert-error">{error}</div>}

          <button className="btn btn-primary btn-block" disabled={busy}>
            {busy
              ? 'Please wait…'
              : mode === 'login'
                ? 'Sign in'
                : 'Create business & account'}
          </button>
        </form>

        {mode === 'login' && (
          <p className="auth-hint">
            Demo: Rajesh / 9876543210 / password123 · Technician Kumar / 9876543211 / password123
          </p>
        )}
      </div>
    </div>
  )
}