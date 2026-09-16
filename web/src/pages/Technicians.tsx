import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api, errorMessage } from '../api/client'
import type { Technician } from '../api/types'
import { Alert, Card, EmptyState, Field, PageHeader, Spinner, StatusBadge } from '../components/ui'

export function TechniciansPage() {
  const qc = useQueryClient()
  const navigate = useNavigate()

  const technicians = useQuery({
    queryKey: ['technicians', 'list'],
    queryFn: async () => (await api.get<{ data: Technician[] }>('/api/v1/technicians')).data,
  })

  const [creating, setCreating] = useState(false)
  const [name, setName] = useState('')
  const [phone, setPhone] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')

  const createTech = useMutation({
    mutationFn: async () =>
      api.post<Technician>('/api/v1/technicians', { name, phone, password }),
    onSuccess: (res) => {
      setCreating(false)
      setName('')
      setPhone('')
      setPassword('')
      qc.invalidateQueries({ queryKey: ['technicians'] })
      navigate(`/technicians/${res.data.id}`)
    },
    onError: (err) => setError(errorMessage(err)),
  })

  return (
    <div>
      <PageHeader
        title="Technicians"
        subtitle="Your field service team"
        action={
          <button className="btn btn-primary" onClick={() => setCreating((v) => !v)}>
            {creating ? 'Close' : 'Add technician'}
          </button>
        }
      />

      {creating && (
        <Card title="Add technician">
          <form
            className="stack"
            onSubmit={(e) => {
              e.preventDefault()
              setError('')
              createTech.mutate()
            }}
          >
            <div className="grid-2">
              <Field label="Name" required>
                <input className="input" required value={name} onChange={(e) => setName(e.target.value)} />
              </Field>
              <Field label="Phone" required>
                <input className="input" required value={phone} onChange={(e) => setPhone(e.target.value)} />
              </Field>
              <Field label="Temporary password" required>
                <input
                  className="input"
                  required
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="At least 6 characters"
                />
              </Field>
            </div>
            <p className="muted">The technician signs in with this phone and password in the mobile app.</p>
            {error && <Alert kind="error">{error}</Alert>}
            <button className="btn btn-primary" disabled={createTech.isPending}>
              {createTech.isPending ? 'Creating…' : 'Create technician'}
            </button>
          </form>
        </Card>
      )}

      {technicians.isLoading ? (
        <Spinner />
      ) : technicians.data && technicians.data.data.length > 0 ? (
        <Card>
          <table className="table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Phone</th>
                <th>Status</th>
                <th>Last login</th>
              </tr>
            </thead>
            <tbody>
              {technicians.data.data.map((t) => (
                <tr key={t.id}>
                  <td>
                    <Link to={`/technicians/${t.id}`} className="link">
                      {t.name}
                    </Link>
                  </td>
                  <td>{t.phone}</td>
                  <td>
                    <StatusBadge value={t.status} />
                  </td>
                  <td>{t.last_login_at ? new Date(t.last_login_at).toLocaleString('en-IN') : 'Never'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      ) : (
        <Card>
          <EmptyState message="No technicians yet. Add your first technician." />
        </Card>
      )}
    </div>
  )
}